FROM golang:1.15-alpine AS builder
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
RUN apk add make gcc musl-dev linux-headers
WORKDIR /app
RUN chown -R appuser:appgroup /app
USER appuser
# User from here
COPY --chown=appuser:appgroup go.* Makefile ./
RUN --mount=type=cache,target=/root/.cache/go-build make dep
COPY --chown=appuser:appgroup . .
#disable tests for now, due to dind which is not available during build
RUN --mount=type=cache,target=/root/.cache/go-build make build

FROM alpine:3.13
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app
COPY --from=builder /app/payout  /app/banner.txt ./
RUN chown -R appuser:appgroup /app
USER appuser
ENTRYPOINT ["./payout"]
