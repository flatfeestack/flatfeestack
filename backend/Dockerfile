FROM golang:1.16 AS base
WORKDIR /app
COPY go.* Makefile ./
RUN make dep

FROM base as builder
COPY *.go *.sql banner.txt ./
RUN make build

FROM gcr.io/distroless/static
WORKDIR /home/nonroot
COPY --from=builder /app/banner.txt /app/backend /app/init.sql ./
USER nonroot
ENTRYPOINT ["/home/nonroot/backend"]
