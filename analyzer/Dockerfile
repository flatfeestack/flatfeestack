FROM golang:1.16 AS base
WORKDIR /app
COPY go.* Makefile ./
RUN make dep

FROM base as builder
COPY *.go ./
RUN make build
RUN ls -la

FROM gcr.io/distroless/static
WORKDIR /home/nonroot
COPY --from=builder /app/analysis-engine ./
USER nonroot
ENTRYPOINT ["/home/nonroot/analysis-engine"]
