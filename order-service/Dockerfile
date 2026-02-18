FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY pkg/ ./pkg/
COPY order-service/go.mod order-service/go.sum ./order-service/
WORKDIR /build/order-service
RUN go mod download

COPY order-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /app ./cmd/main.go

FROM alpine:latest

COPY --from=builder /app /app

EXPOSE 8003

CMD ["/app"]
