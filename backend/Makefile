NAME:=api

.PHONY: all test build clean swag

all: swag dep build test
test:
	go test -v
dep: go.mod
	go mod tidy
	go get -v -u ./...
swag:
	go get -u github.com/swaggo/swag/cmd/swag
	swag init -g main.go
build:
	go build
clean:
	go clean
	rm -f $(NAME)
