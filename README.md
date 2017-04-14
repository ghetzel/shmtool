# shmtool [![GoDoc](https://godoc.org/github.com/ghetzel/shmtool?status.svg)](https://godoc.org/github.com/ghetzel/shmtool/shm)

A command line utility and library for interacting with System V-style shared memory segments, written in Golang.

## Overview
Shared memory is an inter-process communication mechanism that allows for multiple, independent processes to access and modify the same portion of system memory for the purpose of sharing data between them.  This library implements a Golang wrapper around the original implementation of this which is present on almost all *NIX systems that implement portions of the UNIX System V feature set.

The use of the calls implemented by this library has largely been supplanted by [POSIX shared memory](http://man7.org/linux/man-pages/man7/shm_overview.7.html) and the [`mmap()`](http://man7.org/linux/man-pages/man2/mmap.2.html) call, but there are some use cases that still require this particular approach to shared memory management.  One notable example is the [MIT-SHM X Server extension](https://www.x.org/releases/X11R7.7/doc/xextproto/shm.html) which expects SysV shared memory semantics.

## Basic Usage

```golang
package main

import (
  "github.com/ghetzel/shmtool/shm"
  "io"
  "os"
)

func main() {
  // create a 28MiB shared memory segment that other processes can write to
  if segment, err := shm.Create(1024 * 1024 * 28); err == nil {
    // Mark the segment for destruction when the program exits
    //
    // NOTE: The memory segment will only be destroyed when all processes that have attached
    //       to it have detached, which can happen explicitly (here via segment.Detach(ptr)),
    //       or implicitly when the process exits.
    //
    // NOTE: Memory is not overwritten / zeroed out when destroyed.  If you have sensitive data in
    //       this memory segment, you must overwrite it yourself before detaching.
    //
    defer segment.Destroy()

    // Call the Attach() function on the created segment to get the memory address
    // where data can be read or written.  You must do this even if you don't use the address
    // directly as this is what makes the shared memory segment "part of" this process's memory
    // space, and thus allowing you to read from/write to it.
    if segmentAddress, err := segment.Attach(); err == nil {
      defer segment.Detach(segmentAddress)

      // Write the contents of standard input to the shared memory area.
      if _, err := io.Copy(segment, os.Stdin); err != nil {
        panic(err.Error())
      }

      // Do something, maybe tell another process to start and read from this segment (which
      // is communicated by giving the other process the address in segmentAddress).
      //

      // Read the contents of the shared memory area, which may (or may not) have been modified by
      // another program.
      if _, err := io.Copy(os.Stdout, segment); err != nil {
        panic(err.Error())
      }
    } else {
      panic(err.Error())
    }
  } else {
    panic(err.Error())
  }
}
```

## See Also

* [System V interprocess communication mechanisms](http://man7.org/linux/man-pages/man7/svipc.7.html)
* [Use of the `shmget()` system call](http://man7.org/linux/man-pages/man2/shmget.2.html)

