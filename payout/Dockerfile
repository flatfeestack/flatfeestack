FROM golang:1.19 AS base
WORKDIR /app
COPY go.* Makefile ./
RUN make dep

FROM base as builder
COPY *.go banner.txt ./
COPY PayoutNeo.nef PayoutNeo.manifest.json ./
RUN make build

FROM gcr.io/distroless/static
WORKDIR /home/nonroot
COPY --from=builder /app/banner.txt /app/PayoutNeo.nef /app/PayoutNeo.manifest.json /app/payout ./
USER nonroot
ENTRYPOINT ["/home/nonroot/payout"]
