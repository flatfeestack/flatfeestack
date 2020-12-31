all: dep build
dep: go.mod
	go mod tidy
	go get -v -u ./...
build:
	go build
