FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY pkg/ ./pkg/
COPY product-service/go.mod product-service/go.sum ./product-service/
WORKDIR /build/product-service
RUN go mod download

COPY product-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /app ./cmd/main.go

FROM alpine:latest

COPY --from=builder /app /app

EXPOSE 8002

CMD ["/app"]
