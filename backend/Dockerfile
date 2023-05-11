FROM golang:1.20-alpine AS base
WORKDIR /app
COPY go.* ./
RUN go mod download

FROM base as builder

COPY api/ ./api/
COPY clients/ ./clients/
COPY db/ ./db/
COPY utils/ ./utils/
COPY prometheus/ ./prometheus/
COPY *.go *.sql banner.txt ./

RUN go build

FROM alpine:3.18
RUN addgroup -S nonroot -g 31323 && adduser -S nonroot -G nonroot -u 31323
WORKDIR /app
COPY --from=builder /app/banner.txt /app/backend /app/db/init.sql ./
COPY --from=builder /app/db/init.sql ./db/
COPY mail-templates ./mail-templates/
USER nonroot
ENTRYPOINT ["/app/backend"]
