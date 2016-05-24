package shm

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"
)

var segmentId int
var data []byte

func benchmarkAllocateAndDestroy(size int, b *testing.B) {
	for n := 0; n < b.N; n++ {
		segment, _ := Create(size)
		segmentId = segment.Id
		segment.Destroy()
	}
}

func BenchmarkAllocate_1B(b *testing.B)       { benchmarkAllocateAndDestroy(1, b) }
func BenchmarkAllocate_1KB(b *testing.B)      { benchmarkAllocateAndDestroy(1024, b) }
func BenchmarkAllocate_4KB(b *testing.B)      { benchmarkAllocateAndDestroy(4096, b) }
func BenchmarkAllocate_1MB(b *testing.B)      { benchmarkAllocateAndDestroy(1048576, b) }
func BenchmarkAllocate_Buf1080p(b *testing.B) { benchmarkAllocateAndDestroy(2073600, b) }
func BenchmarkAllocate_Buf4KUHD(b *testing.B) { benchmarkAllocateAndDestroy(8294400, b) }
func BenchmarkAllocate_10MB(b *testing.B)     { benchmarkAllocateAndDestroy(10485760, b) }
func BenchmarkAllocate_100MB(b *testing.B)    { benchmarkAllocateAndDestroy(104857600, b) }
func BenchmarkAllocate_1GB(b *testing.B)      { benchmarkAllocateAndDestroy(1073741824, b) }

// Full Read: ioutil
func benchmarkReadFullAuto(size int, b *testing.B) {
	segment, _ := Create(size)
	segmentId = segment.Id

	for n := 0; n < b.N; n++ {
		segment.Reset()
		ioutil.ReadAll(segment)
	}

	segment.Destroy()
}

func BenchmarkReadFullAuto_1B(b *testing.B)       { benchmarkReadFullAuto(1, b) }
func BenchmarkReadFullAuto_1KB(b *testing.B)      { benchmarkReadFullAuto(1024, b) }
func BenchmarkReadFullAuto_4KB(b *testing.B)      { benchmarkReadFullAuto(4096, b) }
func BenchmarkReadFullAuto_1MB(b *testing.B)      { benchmarkReadFullAuto(1048576, b) }
func BenchmarkReadFullAuto_Buf1080p(b *testing.B) { benchmarkReadFullAuto(2073600, b) }
func BenchmarkReadFullAuto_Buf4KUHD(b *testing.B) { benchmarkReadFullAuto(8294400, b) }
func BenchmarkReadFullAuto_10MB(b *testing.B)     { benchmarkReadFullAuto(10485760, b) }
func BenchmarkReadFullAuto_100MB(b *testing.B)    { benchmarkReadFullAuto(104857600, b) }
func BenchmarkReadFullAuto_1GB(b *testing.B)      { benchmarkReadFullAuto(1073741824, b) }

// Full Read: Preallocated Slice
func benchmarkReadFullPreallocate(size int, b *testing.B) {
	segment, _ := Create(size)
	segmentId = segment.Id
	data = make([]byte, size)

	for n := 0; n < b.N; n++ {
		buffer := bytes.NewBuffer(data)
		segment.Reset()
		io.CopyN(buffer, segment, int64(size))
	}

	segment.Destroy()
}

func BenchmarkReadFullPreallocate_1B(b *testing.B)       { benchmarkReadFullPreallocate(1, b) }
func BenchmarkReadFullPreallocate_1KB(b *testing.B)      { benchmarkReadFullPreallocate(1024, b) }
func BenchmarkReadFullPreallocate_4KB(b *testing.B)      { benchmarkReadFullPreallocate(4096, b) }
func BenchmarkReadFullPreallocate_1MB(b *testing.B)      { benchmarkReadFullPreallocate(1048576, b) }
func BenchmarkReadFullPreallocate_Buf1080p(b *testing.B) { benchmarkReadFullPreallocate(2073600, b) }
func BenchmarkReadFullPreallocate_Buf4KUHD(b *testing.B) { benchmarkReadFullPreallocate(8294400, b) }
func BenchmarkReadFullPreallocate_10MB(b *testing.B)     { benchmarkReadFullPreallocate(10485760, b) }
func BenchmarkReadFullPreallocate_100MB(b *testing.B)    { benchmarkReadFullPreallocate(104857600, b) }
func BenchmarkReadFullPreallocate_1GB(b *testing.B)      { benchmarkReadFullPreallocate(1073741824, b) }

// Full Read: Preallocated Slice
func benchmarkReadChunkFull(size int, b *testing.B) {
	segment, _ := Create(size)
	segmentId = segment.Id
	var data []byte

	for n := 0; n < b.N; n++ {
		d, _ := segment.ReadChunk(-1, 0)
		data = d
	}

	if len(data) < size {
		b.Errorf("Expected %d, got: %d", size, len(data))
	}

	segment.Destroy()
}

func BenchmarkReadChunk_1B(b *testing.B)       { benchmarkReadChunkFull(1, b) }
func BenchmarkReadChunk_1KB(b *testing.B)      { benchmarkReadChunkFull(1024, b) }
func BenchmarkReadChunk_4KB(b *testing.B)      { benchmarkReadChunkFull(4096, b) }
func BenchmarkReadChunk_1MB(b *testing.B)      { benchmarkReadChunkFull(1048576, b) }
func BenchmarkReadChunk_Buf1080p(b *testing.B) { benchmarkReadChunkFull(2073600, b) }
func BenchmarkReadChunk_Buf4KUHD(b *testing.B) { benchmarkReadChunkFull(8294400, b) }
func BenchmarkReadChunk_10MB(b *testing.B)     { benchmarkReadChunkFull(10485760, b) }
func BenchmarkReadChunk_100MB(b *testing.B)    { benchmarkReadChunkFull(104857600, b) }
func BenchmarkReadChunk_1GB(b *testing.B)      { benchmarkReadChunkFull(1073741824, b) }
