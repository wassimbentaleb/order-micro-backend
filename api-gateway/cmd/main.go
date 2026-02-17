package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/hero/microservice/api-gateway/internal/routes"
)

func main() {
	cfg := routes.ServiceConfig{
		UserServiceURL:         getEnv("USER_SERVICE_URL", "http://localhost:8001"),
		ProductServiceURL:      getEnv("PRODUCT_SERVICE_URL", "http://localhost:8002"),
		OrderServiceURL:        getEnv("ORDER_SERVICE_URL", "http://localhost:8003"),
		NotificationServiceURL: getEnv("NOTIFICATION_SERVICE_URL", "http://localhost:8004"),
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"service": "api-gateway", "status": "running"})
	})

	routes.SetupRoutes(mux, cfg)

	port := getEnv("SERVER_PORT", "8080")
	log.Printf("API Gateway starting on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
