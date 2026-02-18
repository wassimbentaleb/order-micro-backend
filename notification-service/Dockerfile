FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY pkg/ ./pkg/
COPY notification-service/go.mod notification-service/go.sum ./notification-service/
WORKDIR /build/notification-service
RUN go mod download

COPY notification-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /app ./cmd/main.go

FROM alpine:latest

COPY --from=builder /app /app

EXPOSE 8004

CMD ["/app"]
