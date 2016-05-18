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

// SHM-backed file
type Segment struct {
	Id     int
	Size   int
	offset int
}

// this is where shmget() with IPC_CREAT will happen
func Create(size int) (*Segment, error) {
	return OpenSegment(size, (IpcCreate | IpcExclusive), 0600)
}

// shmget() without IPC_CREAT
func Open(id int) (*Segment, error) {
	if sz, err := C.sysv_shm_get_size(C.int(id)); err == nil {
		return &Segment{
			Id:   id,
			Size: int(sz),
		}, nil
	} else {
		return nil, err
	}
}

func OpenSegment(size int, flags SharedMemoryFlags, perms os.FileMode) (*Segment, error) {
	if shmid, err := C.sysv_shm_open(C.int(size), C.int(flags), C.int(perms)); err == nil {
		if actual_size, err := C.sysv_shm_get_size(shmid); err != nil {
			return nil, fmt.Errorf("Failed to retrieve SHM size: %v", err)
		} else {
			return &Segment{
				Id:   int(shmid),
				Size: int(actual_size),
			}, nil
		}

	} else {
		return nil, err
	}
}

func DestroySegment(id int) error {
	_, err := C.sysv_shm_close(C.int(id))
	return err
}

// will do a memcpy() of len(p) into p from self.addr
func (self *Segment) Read(p []byte) (n int, err error) {
	if self.Id == 0 {
		return 0, fmt.Errorf("Cannot read shared memory segment: SHMID not set")
	}

	// if the offset runs past the segment size, we've reached the end
	if self.offset >= self.Size {
		return 0, io.EOF
	}

	length := len(p)

	// read length cannot exceed segment size
	if length > self.Size {
		length = self.Size
	}

	// if length+offset would overrun, make length equal (size - offset), which is what remains
	if (length + self.offset) > self.Size {
		length = self.Size - self.offset
	}

	buffer := C.malloc(C.size_t(length))

	if _, err := C.sysv_shm_read(C.int(self.Id), buffer, C.int(length), C.int(self.offset)); err != nil {
		return 0, err
	}

	if v := copy(p, C.GoBytes(buffer, C.int(length))); v > 0 {
		self.offset += v
		return v, nil
	} else {
		return v, io.EOF
	}
}

// will do a memcpy() of up to self.size from p to self.addr
func (self *Segment) Write(p []byte) (n int, err error) {
	// if the offset runs past the segment size, we've reached the end
	if self.offset >= self.Size {
		return 0, io.EOF
	}

	length := len(p)

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
		return length, nil
	}
}

func (self *Segment) Reset() {
	self.offset = 0
}

func (self *Segment) Seek(position int) {
	self.offset = position
}

func (self *Segment) Position() int {
	return self.offset
}

func (self *Segment) Attach() (unsafe.Pointer, error) {
	if addr, err := C.sysv_shm_attach(C.int(self.Id)); err == nil {
		return unsafe.Pointer(addr), nil
	} else {
		return nil, err
	}
}

func (self *Segment) Detach(addr unsafe.Pointer) error {
	_, err := C.sysv_shm_detach(addr)
	return err
}
