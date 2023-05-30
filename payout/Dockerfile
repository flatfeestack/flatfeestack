FROM golang:1.20-alpine AS builder

RUN apk add --update --no-cache gcc musl-dev

WORKDIR /app

COPY go.* ./

RUN go mod download

COPY contracts/ ./contracts/
COPY metrics/ ./metrics/
COPY *.go banner.txt PayoutNeo.nef PayoutNeo.manifest.json ./

RUN go build

FROM alpine:3.18

WORKDIR /app

RUN addgroup -S nonroot -g 31323 && adduser -S nonroot -G nonroot -u 31323
RUN apk add --update --no-cache curl

COPY --from=builder /app/banner.txt /app/PayoutNeo.nef /app/PayoutNeo.manifest.json /app/payout ./

USER nonroot

HEALTHCHECK --interval=30s --timeout=3s \
  CMD curl -f http://localhost:${PORT}/metrics || exit 1

ENTRYPOINT ["/app/payout"]
