FROM golang:1.20-alpine AS builder

RUN apk update && apk add --update make gcc musl-dev

WORKDIR /app

COPY go.* ./

RUN go mod download

COPY *.go *.sql login.html banner.txt ./

RUN go build

FROM alpine:3.18

WORKDIR /app

RUN apk add --update --no-cache curl

COPY --from=builder /app/login.html /app/banner.txt /app/fastauth /app/rmdb.sql /app/init.sql ./
COPY mail-templates ./mail-templates/

HEALTHCHECK --interval=30s --timeout=3s \
  CMD curl -f http://localhost:${PORT}/metrics || exit 1

ENTRYPOINT ["/app/fastauth"]
