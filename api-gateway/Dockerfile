FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY pkg/ ./pkg/
COPY api-gateway/go.mod api-gateway/go.sum ./api-gateway/
WORKDIR /build/api-gateway
RUN go mod download

COPY api-gateway/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /app ./cmd/main.go

FROM alpine:latest

COPY --from=builder /app /app

EXPOSE 8080

CMD ["/app"]
