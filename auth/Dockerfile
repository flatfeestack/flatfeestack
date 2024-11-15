FROM golang:1.23-alpine AS base
RUN apk update && apk add --update make gcc musl-dev
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY *.go *.sql login.html banner.txt ./
RUN go build

FROM alpine:3.20
WORKDIR /app
RUN addgroup -S nonroot -g 31323 && adduser -S nonroot -G nonroot -u 31323
COPY --from=base /app/login.html /app/banner.txt /app/auth /app/rmdb.sql /app/init.sql ./
COPY mail-templates ./mail-templates/
USER nonroot
HEALTHCHECK --interval=30s --timeout=3s CMD wget --quiet --tries=1 --spider http://localhost:${PORT}/metrics || exit 1
ENTRYPOINT ["/app/auth"]
