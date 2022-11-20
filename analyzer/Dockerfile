FROM golang:1.19-alpine AS builder
RUN apk add --update --no-cache cmake git make gcc musl-dev openssl-dev
WORKDIR /app
RUN git clone https://github.com/libgit2/libgit2.git
RUN cd /app/libgit2 && git checkout v1.3.0 && mkdir build && cd build && cmake .. -DCMAKE_INSTALL_PREFIX=/usr -DBUILD_CLAR=OFF && cmake --build . --target install -j 8
COPY go.* ./
RUN go mod download
COPY *.go banner.txt ./
RUN go build

FROM alpine:3.16
RUN apk add --update --no-cache git
RUN addgroup -S nonroot && adduser -S nonroot -G nonroot
WORKDIR /app
COPY --from=builder /app/banner.txt /app/analysis-engine ./
COPY --from=builder /usr/lib/libgit* /usr/lib/
#USER nonroot
#https://stackoverflow.com/questions/19331497/set-environment-variables-from-file-of-key-value-pairs/30969768#30969768
ENTRYPOINT ["/app/analysis-engine"]
