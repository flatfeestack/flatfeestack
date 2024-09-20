FROM golang:1.23-alpine AS base
WORKDIR /app
COPY go.* ./
RUN go mod download

FROM base as builder
COPY --from=base /app/*.go ./
COPY internal/ ./internal/
COPY cmd/ ./cmd/
COPY migrations/ ./migrations/
COPY pkg/ ./pkg/
COPY templates/ ./templates/
WORKDIR /app/cmd
RUN go build -o /app/backend

FROM alpine:3.20
WORKDIR /app
RUN addgroup -S nonroot -g 31323 && adduser -S nonroot -G nonroot -u 31323
RUN apk add --update --no-cache curl
COPY --from=builder /app/migrations/init.sql /app/cmd/banner.txt /app/backend ./
COPY --from=builder /app/templates ./templates
USER nonroot
HEALTHCHECK --interval=30s --timeout=3s CMD curl -f http://localhost:${PORT}/metrics || exit 1
ENTRYPOINT ["/app/backend"]
