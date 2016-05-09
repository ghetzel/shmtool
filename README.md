# shmtool [![GoDoc](https://godoc.org/github.com/ghetzel/shmtool?status.svg)](https://godoc.org/github.com/ghetzel/shmtool)

A command line utility and library for interacting with System V-style shared memory segments, written in Golang.

## Overview
Shared memory is an inter-process communication mechanism that allows for multiple, independent processes to access and modify the same portion of system memory for the purpose of sharing data between them.  This library implements a Golang wrapper around the original implementation of this which is present on almost all *NIX systems that implement portions of the UNIX System V feature set.

The use of the calls implemented by this library has largely been supplanted by [POSIX shared memory](http://man7.org/linux/man-pages/man7/shm_overview.7.html) and the [`mmap()`](http://man7.org/linux/man-pages/man2/mmap.2.html) call, but there are some use cases that still require this particular approach to shared memory management.  One notable example is the [MIT-SHM X Server extension](https://www.x.org/releases/X11R7.7/doc/xextproto/shm.html) which expects SysV shared memory semantics.

## See Also

* [System V interprocess communication mechanisms](http://man7.org/linux/man-pages/man7/svipc.7.html)
* [Use of the `shmget()` system call](http://man7.org/linux/man-pages/man2/shmget.2.html)

