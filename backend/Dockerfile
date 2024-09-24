FROM golang:1.23-alpine AS base
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY *.go banner.txt ./
COPY internal/ ./internal/
COPY db/init.sql ./
COPY pkg/ ./pkg/
COPY templates/ ./templates/
RUN go build

FROM alpine:3.20
WORKDIR /app
RUN addgroup -S nonroot -g 31323 && adduser -S nonroot -G nonroot -u 31323
COPY --from=base /app/db/init.sql ./db/init.sql
COPY --from=base /app/banner.txt /app/backend ./
COPY --from=base /app/templates ./templates
USER nonroot
HEALTHCHECK --interval=30s --timeout=3s CMD wget --quiet --tries=1 --spider http://localhost:${PORT}/metrics || exit 1
ENTRYPOINT ["/app/backend"]
