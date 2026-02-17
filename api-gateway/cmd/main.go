package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/hero/microservice/api-gateway/internal/routes"
)

func main() {
	cfg := routes.ServiceConfig{
		UserServiceURL:         getEnv("USER_SERVICE_URL", "http://localhost:8001"),
		ProductServiceURL:      getEnv("PRODUCT_SERVICE_URL", "http://localhost:8002"),
		OrderServiceURL:        getEnv("ORDER_SERVICE_URL", "http://localhost:8003"),
		NotificationServiceURL: getEnv("NOTIFICATION_SERVICE_URL", "http://localhost:8004"),
	}

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "api-gateway", "status": "running"})
	})

	routes.SetupRoutes(r, cfg)

	port := getEnv("SERVER_PORT", "8080")
	log.Printf("API Gateway starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
