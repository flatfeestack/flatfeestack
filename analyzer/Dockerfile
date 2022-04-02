FROM golang:1.17-alpine3.15 AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
RUN apk add cmake git make gcc musl-dev openssl-dev
RUN git clone https://github.com/libgit2/libgit2.git
WORKDIR /app/libgit2
RUN git checkout v1.3.0
RUN mkdir build;
WORKDIR /app/libgit2/build
RUN cmake .. -DCMAKE_INSTALL_PREFIX=/usr -DBUILD_CLAR=OFF && cmake --build . && cmake --build . --target install
WORKDIR /app
COPY *.go ./
RUN go build

FROM alpine:3.15
RUN apk add --update --no-cache git
RUN addgroup -S nonroot && adduser -S nonroot -G nonroot
WORKDIR /app
COPY --from=builder /app/analysis-engine ./
COPY banner.txt ./
COPY --from=builder /usr/lib/libgit* /usr/lib/
USER nonroot
ENTRYPOINT ["/app/analysis-engine"]
