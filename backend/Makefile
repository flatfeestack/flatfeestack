.PHONY: all dep build test clean

all: clean dep build test
dep:
	go mod download
build:
	CGO_ENABLED=0 go build
test:
	go test
clean:
	go clean
	go mod tidy
