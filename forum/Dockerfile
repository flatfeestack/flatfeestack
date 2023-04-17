FROM golang:1.20-alpine AS base
WORKDIR /app
COPY go.* ./
RUN go mod download

FROM base as builder
COPY /api/*.go ./api/
COPY /db/*.go ./db/
COPY /globals/*.go ./globals/
COPY /jwt/*.go ./jwt/
COPY /types/*.go ./types/
COPY /utils/*.go ./utils/
COPY *.go *.sql ./
RUN ls -all
RUN go build

FROM alpine:3.17
RUN addgroup -S nonroot -g 31323 && adduser -S nonroot -G nonroot -u 31323
WORKDIR /app
COPY --from=builder /app/forum /app/init.sql ./
USER nonroot
ENTRYPOINT ["/app/forum"]
