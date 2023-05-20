FROM golang:1.20-alpine AS base
WORKDIR /app
COPY go.* ./
RUN go mod download

FROM base as builder
COPY /api/*.go ./api/
COPY /dao/*.go ./dao/
COPY /db/*.go /db/*.sql ./db/
COPY /globals/*.go ./globals/
COPY /jwt/*.go ./jwt/
COPY /types/*.go ./types/
COPY /utils/*.go ./utils/
COPY *.go banner.txt ./

RUN go build

FROM alpine:3.18
RUN addgroup -S nonroot -g 31323 && adduser -S nonroot -G nonroot -u 31323
WORKDIR /app
COPY --from=builder /app/forum /app/db/init.sql /app/banner.txt ./
USER nonroot
ENTRYPOINT ["/app/forum"]
