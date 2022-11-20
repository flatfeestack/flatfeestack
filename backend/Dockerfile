FROM golang:1.19-alpine AS base
WORKDIR /app
COPY go.* ./
RUN go mod download

FROM base as builder
COPY *.go *.sql banner.txt ./
RUN go build

FROM alpine:3.17
WORKDIR /app
RUN addgroup -S nonroot -g 31323 && adduser -S nonroot -G nonroot -u 31323
COPY --from=builder /app/banner.txt /app/backend /app/init.sql ./
USER nonroot
ENTRYPOINT ["/app/backend"]
