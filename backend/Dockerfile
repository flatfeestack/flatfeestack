FROM golang:1.14

ENV GOPATH /go

COPY . /go/src/api
WORKDIR /go/src/api

RUN make

CMD ["./api"]
