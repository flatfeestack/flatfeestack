FROM golang:1.20-alpine AS base
WORKDIR /app
COPY go.* ./
RUN go mod download

FROM base as builder

COPY internal/api/ ./api/
COPY clients/ ./clients/
COPY db/ ./db/
COPY utils/ ./utils/
COPY *.go *.sql banner.txt ./

RUN go build

FROM alpine:3.18

WORKDIR /app

RUN addgroup -S nonroot -g 31323 && adduser -S nonroot -G nonroot -u 31323
RUN apk add --update --no-cache curl

COPY --from=builder /app/banner.txt /app/backend /app/db/init.sql ./
COPY --from=builder /app/db/init.sql ./db/

COPY templates ./mail-templates/

USER nonroot

HEALTHCHECK --interval=30s --timeout=3s \
  CMD curl -f http://localhost:${PORT}/metrics || exit 1

ENTRYPOINT ["/app/backend"]
