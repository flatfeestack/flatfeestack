FROM golang:1.19-alpine AS base
RUN apk add --update --no-cache gcc musl-dev
WORKDIR /app
COPY go.* ./
RUN go mod download

FROM base as builder
COPY *.go banner.txt PayoutNeo.nef PayoutNeo.manifest.json ./
RUN go build

FROM alpine:3.17
RUN addgroup -S nonroot -g 31323 && adduser -S nonroot -G nonroot -u 31323
WORKDIR /app
COPY --from=builder /app/banner.txt /app/PayoutNeo.nef /app/PayoutNeo.manifest.json /app/payout ./
USER nonroot
ENTRYPOINT ["/app/payout"]
