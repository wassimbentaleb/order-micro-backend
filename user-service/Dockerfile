FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY pkg/ ./pkg/
COPY user-service/go.mod user-service/go.sum ./user-service/
WORKDIR /build/user-service
RUN go mod download

COPY user-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /app ./cmd/main.go

FROM alpine:latest

COPY --from=builder /app /app

EXPOSE 8001

CMD ["/app"]
