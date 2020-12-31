FROM golang:1.14 AS builder
WORKDIR /app
COPY . .
RUN make

FROM golang:1.14
WORKDIR /app
COPY --from=builder /app/payout ./
ENTRYPOINT ["./payout"]
