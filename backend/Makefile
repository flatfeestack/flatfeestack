NAME:=fastauth

build:
	go get -u github.com/swaggo/swag/cmd/swag
	swag init -g main.go
	go build
run:
	./api