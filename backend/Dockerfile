FROM golang:1.17 AS base
WORKDIR /app
COPY go.* Makefile ./
RUN make dep

FROM base as builder
COPY *.go *.sql banner.txt ./
RUN make build

FROM gcr.io/distroless/static
WORKDIR /home/nonroot
COPY --from=builder /app/banner.txt /app/backend /app/init.sql /app/init_func.sql ./
USER nonroot
ENTRYPOINT ["/home/nonroot/backend"]
