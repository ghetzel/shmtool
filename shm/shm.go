// Shared memory is an inter-process communication mechanism that allows for multiple,
// independent processes to access and modify the same portion of system memory for the purpose of
// sharing data between them.  This library implements a Golang wrapper around the original
// implementation of this which is present on almost all *NIX systems that implement portions of the
// UNIX System V feature set.
//
// The use of the calls implemented by this library has largely been supplanted by POSIX shared memory
// (http://man7.org/linux/man-pages/man7/shm_overview.7.html) and the mmap() system call, but there are
// some use cases that still require this particular approach to shared memory management.  One notable
// example is the MIT-SHM X Server extension
// (https://www.x.org/releases/X11R7.7/doc/xextproto/shm.html) which expects SysV shared memory
// semantics.
//
package shm

// #include "shm.h"
import "C"

import (
	"fmt"
	"io"
	"os"
	"unsafe"
)

const Version = `0.0.2`

type SharedMemoryFlags int

const (
	IpcNone                        = 0
	IpcCreate    SharedMemoryFlags = C.IPC_CREAT
	IpcExclusive                   = C.IPC_EXCL
	HugePages                      = C.SHM_HUGETLB
	NoReserve                      = C.SHM_NORESERVE
)

// A native representation of a SysV shared memory segment
type Segment struct {
	Id     int
	Size   int64
	offset int64
}

// Create a new shared memory segment with the given size (in bytes).  The system will automatically
// round the size up to the nearest memory page boundary (typically 4KB).
//
func Create(size int) (*Segment, error) {
	return OpenSegment(size, (IpcCreate | IpcExclusive), 0600)
}

// Open an existing shared memory segment located at the given ID.  This ID is returned in the
// struct that is populated by Create(), or by the shmget() system call.
//
func Open(id int) (*Segment, error) {
	if sz, err := C.sysv_shm_get_size(C.int(id)); err == nil {
		return &Segment{
			Id:   id,
			Size: int64(sz),
		}, nil
	} else {
		return nil, err
	}
}

// Creates a shared memory segment of a given size, and also allows for the specification of
// creation flags supported by the shmget() call, as well as specifying permissions.
//
func OpenSegment(size int, flags SharedMemoryFlags, perms os.FileMode) (*Segment, error) {
	if shmid, err := C.sysv_shm_open(C.int(size), C.int(flags), C.int(perms)); err == nil {
		if actual_size, err := C.sysv_shm_get_size(shmid); err != nil {
			return nil, fmt.Errorf("Failed to retrieve SHM size: %v", err)
		} else {
			return &Segment{
				Id:   int(shmid),
				Size: int64(actual_size),
			}, nil
		}

	} else {
		return nil, err
	}
}

// Destroy a shared memory segment by its ID
//
func DestroySegment(id int) error {
	_, err := C.sysv_shm_close(C.int(id))
	return err
}

// Read some or all of the shared memory segment and return a byte slice.
//
func (self *Segment) ReadChunk(length int64, start int64) ([]byte, error) {
	if length < 0 {
		length = self.Size
	}

	buffer := C.malloc(C.size_t(length))
	defer C.free(buffer)

	if _, err := C.sysv_shm_read(C.int(self.Id), buffer, C.int(length), C.int(start)); err != nil {
		return nil, err
	}

	return C.GoBytes(buffer, C.int(length)), nil
}

// Implements the io.Reader interface for shared memory
//
func (self *Segment) Read(p []byte) (n int, err error) {
	if self.Id == 0 {
		return 0, fmt.Errorf("Cannot read shared memory segment: SHMID not set")
	}

	// if the offset runs past the segment size, we've reached the end
	if self.offset >= self.Size {
		return 0, io.EOF
	}

	length := int64(len(p))

	// read length cannot exceed segment size
	if length > self.Size {
		length = self.Size
	}

	// if length+offset would overrun, make length equal (size - offset), which is what remains
	if (length + self.offset) > self.Size {
		length = self.Size - self.offset
	}

	buffer := C.malloc(C.size_t(length))
	defer C.free(buffer)

	if _, err := C.sysv_shm_read(C.int(self.Id), buffer, C.int(length), C.int(self.offset)); err != nil {
		return 0, err
	}

	if v := copy(p, C.GoBytes(buffer, C.int(length))); v > 0 {
		self.offset += int64(v)
		return v, nil
	} else {
		return v, io.EOF
	}
}

// Implements the io.Writer interface for shared memory
//
func (self *Segment) Write(p []byte) (n int, err error) {
	// if the offset runs past the segment size, we've reached the end
	if self.offset >= self.Size {
		return 0, io.EOF
	}

	length := int64(len(p))

	// write length cannot exceed segment size
	if length > self.Size {
		length = self.Size
	}

	// if length+offset would overrun, make length equal (size - offset), which is what remains
	if (length + self.offset) > self.Size {
		length = self.Size - self.offset
	}

	if _, err := C.sysv_shm_write(C.int(self.Id), unsafe.Pointer(&p[0]), C.int(length), C.int(self.offset)); err != nil {
		return 0, err
	} else {
		self.offset += length
		return int(length), nil
	}
}


// Resets the internal offset counter for this segment, allowing subsequent calls
// to Read() or Write() to start from the beginning.
//
func (self *Segment) Reset() {
	self.offset = 0
}

// Implements the io.Seeker interface for shared memory.  Subsequent calls to Read()
// or Write() will start from this position.
//
func (self *Segment) Seek(offset int64, whence int) (int64, error) {
	var computedOffset int64

	switch whence {
	case 1:
		computedOffset = self.offset + offset
	case 2:
		computedOffset = self.Size - offset
	default:
		computedOffset = offset
	}

	if computedOffset < 0 {
		return 0, fmt.Errorf("Cannot seek to position before start of segment")
	}

	self.offset = computedOffset
	return self.offset, nil
}

// Returns the current position of the Read/Write pointer.
//
func (self *Segment) Position() int64 {
	return self.offset
}

// Attaches the segment to the current processes resident memory.  The pointer
// that is returned is the actual memory address of the shared memory segment
// for use with third party libraries that can directly read from memory.
//
func (self *Segment) Attach() (unsafe.Pointer, error) {
	if addr, err := C.sysv_shm_attach(C.int(self.Id)); err == nil {
		return unsafe.Pointer(addr), nil
	} else {
		return nil, err
	}
}

// Detaches the segment from the current processes memory space.
//
func (self *Segment) Detach(addr unsafe.Pointer) error {
	_, err := C.sysv_shm_detach(addr)
	return err
}

// Destroys the current shared memory segment.
//
func (self *Segment) Destroy() error {
	return DestroySegment(self.Id)
}
