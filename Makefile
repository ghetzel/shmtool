.PHONY: fmt build test bench
.EXPORT_ALL_VARIABLES:

GO111MODULE ?= on

all: fmt build test

fmt:
	gofmt -w .
	gofmt -w shm/..

test:
	go test -v ./shm

bench:
	go test -bench=. ./shm

build:
	go build -o bin/shmtool
