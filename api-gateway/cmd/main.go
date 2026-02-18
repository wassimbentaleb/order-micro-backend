package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/hero/microservice/api-gateway/internal/middleware"
	"github.com/hero/microservice/api-gateway/internal/routes"
	"github.com/hero/microservice/pkg/cache"
)

func main() {
	cfg := routes.ServiceConfig{
		UserServiceURL:         getEnv("USER_SERVICE_URL", "http://localhost:8001"),
		ProductServiceURL:      getEnv("PRODUCT_SERVICE_URL", "http://localhost:8002"),
		OrderServiceURL:        getEnv("ORDER_SERVICE_URL", "http://localhost:8003"),
		NotificationServiceURL: getEnv("NOTIFICATION_SERVICE_URL", "http://localhost:8004"),
	}

	// Redis
	rdb, err := cache.NewRedisClient(
		getEnv("REDIS_HOST", "localhost"),
		getEnv("REDIS_PORT", "6379"),
	)
	if err != nil {
		log.Fatal("Failed to connect to Redis: ", err)
	}
	defer rdb.Close()

	// Rate limit config
	rateLimit, _ := strconv.Atoi(getEnv("RATE_LIMIT", "100"))
	rateWindow, _ := strconv.Atoi(getEnv("RATE_WINDOW", "60"))

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"service": "api-gateway", "status": "running"})
	})

	routes.SetupRoutes(mux, cfg, rdb)

	// Apply rate limiting to all routes
	handler := middleware.RateLimitMiddleware(rdb, rateLimit, time.Duration(rateWindow)*time.Second)(mux)

	port := getEnv("SERVER_PORT", "8080")
	log.Printf("API Gateway starting on port %s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
