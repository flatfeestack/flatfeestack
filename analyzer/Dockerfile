FROM golang:1.16-alpine AS base
RUN apk update && apk add --update make gcc musl-dev
WORKDIR /app
COPY go.* Makefile ./
RUN make dep

FROM base as builder
COPY *.go ./
RUN make clean build

FROM gcr.io/distroless/static
WORKDIR /home/nonroot
COPY --from=builder /app/analysis-engine ./
USER nonroot
ENTRYPOINT ["/home/nonroot/analysis-engine"]
