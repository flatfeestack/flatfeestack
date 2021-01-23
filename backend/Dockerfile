#syntax=docker/dockerfile:1.2

#https://www.docker.com/blog/containerize-your-go-developer-environment-part-2/
#https://github.com/moby/buildkit/blob/master/frontend/dockerfile/docs/syntax.md
FROM golang:1.15-alpine AS builder
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
RUN apk add make gcc musl-dev
WORKDIR /app
RUN chown -R appuser:appgroup /app
USER appuser
# User from here
COPY --chown=appuser:appgroup go.* Makefile ./
RUN --mount=type=cache,target=/root/.cache/go-build make dep
COPY --chown=appuser:appgroup . .
#disable tests for now, due to dind which is not available during build
RUN --mount=type=cache,target=/root/.cache/go-build make build swag

FROM alpine:3.13
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app
COPY --from=builder /app/backend  /app/banner.txt ./
RUN chown -R appuser:appgroup /app
USER appuser
ENTRYPOINT ["./backend"]
