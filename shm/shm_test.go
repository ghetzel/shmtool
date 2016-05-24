package shm

import (
    "testing"
    "hash/adler32"
    "bytes"
    "io"
    "io/ioutil"
    "fmt"
)

func makeSegment(t *testing.T, size int, callback func(segment *Segment) error) {
	segment, err := Create(1024)
	defer segment.Destroy()

	if err != nil {
		t.Errorf("Failed to allocate 1024b segment: %v", err)
	}else{
		if err := callback(segment); err != nil {
			t.Error(err)
		}
	}
}

func writeFullSegment(t *testing.T, size int, callback func(segment *Segment, input []byte) error) {
	makeSegment(t, 1024, func(segment *Segment) error {
		input := make([]byte, 1024)
		for i := 0; i < len(input); i++ {
			input[i] = byte(i % 256)
		}

		if n, err := segment.Write(input); err == nil {
			if n != len(input) {
				return fmt.Errorf("Incorrect write size; expected: %d, was: %d", len(input), n)
			}

			segment.Reset()

			return callback(segment, input)
		}else{
			return fmt.Errorf("Failed to write segment data: %v", err)
		}
	})
}

func TestAllocate(t *testing.T) {
	makeSegment(t, 1024, func(segment *Segment) error {
		return nil
	})
}

func TestWriteFull(t *testing.T) {
	writeFullSegment(t, 1024, func(segment *Segment, input []byte) error {
		shouldBe := adler32.Checksum(input)

		// read back and make sure it's correct
		if output, err := ioutil.ReadAll(segment); err == nil {
			if len(output) != len(input) {
				return fmt.Errorf("Incorrect readback size; expected: %d, was: %d", len(input), len(output))
			}

			actuallyIs := adler32.Checksum(output)

			if shouldBe != actuallyIs {
				return fmt.Errorf("Checksum of output does not match input; expected: %d, got: %d", shouldBe, actuallyIs)
			}else{
				t.Logf("Checksum OK: input[%d] %d == output[%d] %d", len(input), shouldBe, len(output), actuallyIs)
			}
		}

		return nil
	})
}

func TestWriteFullPartialReadHead(t *testing.T) {
	writeFullSegment(t, 1024, func(segment *Segment, input []byte) error {
		shouldBe := adler32.Checksum(input[0:512])

		var outwriter bytes.Buffer

		// read back first 512b and make sure it's correct
		if _, err := io.CopyN(&outwriter, segment, 512); err == nil {
			output := outwriter.Bytes()

			if len(output) != 512 {
				return fmt.Errorf("Incorrect readback size; expected: %d, was: %d", 512, len(output))
			}

			actuallyIs := adler32.Checksum(output)

			if shouldBe != actuallyIs {
				return fmt.Errorf("Checksum of output does not match input; expected: %d, got: %d", shouldBe, actuallyIs)
			}else{
				t.Logf("Checksum OK: input[0:512] %d == output[%d] %d", shouldBe, len(output), actuallyIs)
			}
		}

		return nil
	})
}


func TestWriteFullPartialReadTail(t *testing.T) {
	writeFullSegment(t, 1024, func(segment *Segment, input []byte) error {
		shouldBe := adler32.Checksum(input[512:1024])

		// read back first 512b and make sure it's correct
		segment.Seek(512)

		if output, err := ioutil.ReadAll(segment); err == nil {
			if len(output) != 512 {
				return fmt.Errorf("Incorrect readback size; expected: %d, was: %d", 512, len(output))
			}

			actuallyIs := adler32.Checksum(output)

			if shouldBe != actuallyIs {
				return fmt.Errorf("Checksum of output does not match input; expected: %d, got: %d", shouldBe, actuallyIs)
			}else{
				t.Logf("Checksum OK: input[512:] %d == output[%d] %d", shouldBe, len(output), actuallyIs)
			}
		}

		return nil
	})
}

func TestWriteFullPartialReadMiddle(t *testing.T) {
	writeFullSegment(t, 1024, func(segment *Segment, input []byte) error {
		shouldBe := adler32.Checksum(input[256:768])

		var outwriter bytes.Buffer

		// read back first 512b and make sure it's correct
		segment.Seek(256)

		if _, err := io.CopyN(&outwriter, segment, 512); err == nil {
			output := outwriter.Bytes()

			if len(output) != 512 {
				return fmt.Errorf("Incorrect readback size; expected: %d, was: %d", 512, len(output))
			}

			actuallyIs := adler32.Checksum(output)

			if shouldBe != actuallyIs {
				return fmt.Errorf("Checksum of output does not match input; expected: %d, got: %d", shouldBe, actuallyIs)
			}else{
				t.Logf("Checksum OK: input[256:768] %d == output[%d] %d", shouldBe, len(output), actuallyIs)
			}
		}

		return nil
	})
}



func TestWriteFullPartialReadChunks(t *testing.T) {
	writeFullSegment(t, 1024, func(segment *Segment, input []byte) error {
		var err error
		output := make([]byte, 4)


		segment.Seek(255)
		_, err = segment.Read(output[0:1])
		if err != nil {
			return err
		}

		segment.Seek(511)
		_, err = segment.Read(output[1:2])
		if err != nil {
			return err
		}

		segment.Seek(767)
		_, err = segment.Read(output[2:3])
		if err != nil {
			return err
		}

		segment.Seek(1023)
		_, err = segment.Read(output[3:4])
		if err != nil {
			return err
		}

		for i, v := range output {
			if v != 0xFF {
				return fmt.Errorf("Wrong value for output[%d]; expected: 0xFF, got: %X", i, v)
			}
		}

		return nil
	})
}

