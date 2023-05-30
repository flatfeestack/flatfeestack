FROM golang:1.20-bullseye AS builder

ENV DEBIAN_FRONTEND=noninteractive \
    PKG_CONFIG_PATH="$PKG_CONFIG_PATH:/opt/libgit2/lib/pkgconfig" \
    LD_LIBRARY_PATH="$LD_LIBRARY_PATH:/opt/libgit2/lib"

RUN apt update && \
    apt install -y --no-install-recommends ca-certificates cmake gcc git g++ libssh2-1-dev libssl-dev make pkg-config && \
    apt clean

WORKDIR /app
RUN git clone https://github.com/libgit2/libgit2.git && \
    cd /app/libgit2 && git checkout v1.5.0 && \
    mkdir build && \
    cd build && \
    cmake .. -DCMAKE_INSTALL_PREFIX=/opt/libgit2 -DBUILD_TESTS=false && \
    cmake --build . --target install -j $(nproc)

COPY go.* ./
RUN go mod download

COPY *.go ./
RUN go build

FROM debian:bullseye

ENV LD_LIBRARY_PATH="$LD_LIBRARY_PATH:/opt/libgit2/lib"

WORKDIR /app

RUN apt update && \
    apt install -y --no-install-recommends ca-certificates curl && \
    apt clean

COPY banner.txt .

COPY --from=builder /opt/libgit2 /opt/libgit2/
COPY --from=builder /app/analyzer .

HEALTHCHECK --interval=30s --timeout=3s \
  CMD curl -f http://localhost:${PORT}/metrics || exit 1

ENTRYPOINT ["/app/analyzer"]
