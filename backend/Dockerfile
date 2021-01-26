#syntax=docker/dockerfile:1.2

#https://www.docker.com/blog/containerize-your-go-developer-environment-part-2/
#https://github.com/moby/buildkit/blob/master/frontend/dockerfile/docs/syntax.md
FROM golang:1.15-alpine AS builder
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
RUN apk add make gcc musl-dev git
#https://stackoverflow.com/questions/29556879/is-it-possible-change-date-in-docker-container
WORKDIR /
RUN git clone https://github.com/wolfcw/libfaketime.git
WORKDIR /libfaketime/src
RUN make install
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
ARG env
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app
COPY --from=builder /app/backend  /app/banner.txt /app/backend.sh ./
COPY --from=builder /usr/local/ /usr/local/
RUN chown -R appuser:appgroup /app && chmod u+x /app/backend.sh
USER appuser
#https://stackoverflow.com/questions/29556879/is-it-possible-change-date-in-docker-container
ENTRYPOINT ["/bin/sh", "-c", "./backend.sh"]
