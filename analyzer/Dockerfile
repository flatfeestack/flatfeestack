FROM golang:1.20-alpine3.17 AS builder
RUN apk add --update --no-cache git cmake make gcc musl-dev openssl-dev
WORKDIR /app
RUN git clone https://github.com/libgit2/libgit2.git
RUN cd /app/libgit2 && git checkout v1.5.0 && mkdir build && cd build && cmake .. -DCMAKE_INSTALL_PREFIX=/usr -DBUILD_TESTS=false && cmake --build . --target install -j 8
COPY go.* ./
RUN go mod download
COPY *.go ./
RUN go build

FROM alpine:3.18
RUN addgroup -S nonroot && adduser -S nonroot -G nonroot
WORKDIR /app
COPY banner.txt .
COPY --from=builder /app/analyzer .
COPY --from=builder /usr/lib/libgit* /usr/lib/
#USER nonroot
#https://stackoverflow.com/questions/19331497/set-environment-variables-from-file-of-key-value-pairs/30969768#30969768
ENTRYPOINT ["/app/analyzer"]
