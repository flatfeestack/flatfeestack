.PHONY: all dep build swag test clean

all: dep build swag test
dep:
	go mod download
	go get -u github.com/swaggo/swag/cmd/swag
build:
	go build
swag:
	swag init -g main.go
test:
	go test
clean:
	go clean
