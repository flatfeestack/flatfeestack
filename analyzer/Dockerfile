FROM golang:1.20-bullseye AS builder

ENV DEBIAN_FRONTEND=noninteractive

RUN apt update && \
    apt install -y --no-install-recommends ca-certificates cmake gcc git g++ libssh2-1-dev libssl-dev make pkg-config && \
    apt clean

WORKDIR /app
RUN git clone https://github.com/libgit2/libgit2.git && \
    cd /app/libgit2 && git checkout v1.5.0 && \
    mkdir build && \
    cd build && \
    cmake .. -DCMAKE_INSTALL_PREFIX=/usr -DBUILD_TESTS=false && \
    cmake --build . --target install -j $(nproc)

COPY go.* ./
RUN go mod download

COPY *.go ./
RUN go build

FROM debian:bullseye

WORKDIR /app

RUN apt update && \
    apt install -y --no-install-recommends ca-certificates && \
    apt clean

COPY banner.txt .

COPY --from=builder /usr/lib/x86_64-linux-gnu/libgit2* /usr/lib/x86_64-linux-gnu/
COPY --from=builder /usr/lib/x86_64-linux-gnu/pkgconfig/libgit2.pc /usr/lib/x86_64-linux-gnu/pkgconfig/libgit2.pc
COPY --from=builder /usr/include/git2 /usr/include/git2/

COPY --from=builder /app/analyzer .

ENTRYPOINT ["/app/analyzer"]
