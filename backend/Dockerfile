FROM golang:1.15-alpine AS builder
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
RUN apk add make gcc musl-dev
WORKDIR /app
COPY . .
RUN chown -R appuser:appgroup /app
USER appuser
#disable tests for now, due to dind which is not available during build
RUN make swag dep build

FROM alpine:3.13
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app
COPY --from=builder /app/backend  /app/banner.txt ./
RUN chown -R appuser:appgroup /app
USER appuser
ENTRYPOINT ["./backend"]
