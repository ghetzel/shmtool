all: vendor fmt build

update:
	rm -rf vendor
	glide up --strip-vcs --update-vendored

vendor:
	go list github.com/Masterminds/glide
	glide install --strip-vcs --update-vendored

fmt:
	gofmt -w .
	gofmt -w shm/..

build:
	go build -o bin/`basename ${PWD}`
