FROM golang:1.17 AS base
WORKDIR /app
COPY go.* ./
RUN go mod download

FROM base as builder
WORKDIR /app
RUN apt update && apt install -y cmake libssl-dev
RUN git clone https://github.com/libgit2/libgit2.git
WORKDIR /app/libgit2
RUN git checkout v1.3.0
RUN mkdir build;
WORKDIR /app/libgit2/build
RUN cmake ..
RUN cmake --build .
RUN cmake --build . --target install
WORKDIR /app
COPY *.go ./
RUN go build

FROM gcr.io/distroless/static
WORKDIR /home/nonroot
COPY --from=builder /app/analysis-engine ./
COPY --from=builder /usr/local/lib/libgit* /usr/local/lib/
USER nonroot
ENTRYPOINT ["/home/nonroot/analysis-engine"]
