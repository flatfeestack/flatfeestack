FROM golang:1.20-alpine AS base
WORKDIR /app
COPY go.* ./
RUN go mod download

FROM base as builder
COPY *.go *.sql banner.txt ./
COPY mail-templates ./mail-templates/
RUN go build

FROM alpine:3.17
RUN addgroup -S nonroot -g 31323 && adduser -S nonroot -G nonroot -u 31323
WORKDIR /app
COPY --from=builder /app/banner.txt /app/backend /app/init.sql ./
COPY --from=builder /app/mail-templates/ ./mail-templates/
USER nonroot
ENTRYPOINT ["/app/backend"]
