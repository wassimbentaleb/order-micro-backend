package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/hero/microservice/order-service/internal/handler"
	"github.com/hero/microservice/order-service/internal/rabbitmq"
	"github.com/hero/microservice/order-service/internal/repository"
	"github.com/hero/microservice/order-service/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s search_path=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "svc_order"),
		getEnv("DB_PASSWORD", "svc_order_pass"),
		getEnv("DB_NAME", "microservice_db"),
		getEnv("DB_SCHEMA", "order_schema"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}
	log.Println("Database connected")

	// RabbitMQ publisher
	publisher, err := rabbitmq.NewPublisher(
		getEnv("RABBITMQ_HOST", "localhost"),
		getEnv("RABBITMQ_PORT", "5672"),
		getEnv("RABBITMQ_USER", "guest"),
		getEnv("RABBITMQ_PASSWORD", "guest"),
	)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ: ", err)
	}
	defer publisher.Close()

	// RabbitMQ consumer
	consumer, err := rabbitmq.NewConsumer(
		getEnv("RABBITMQ_HOST", "localhost"),
		getEnv("RABBITMQ_PORT", "5672"),
		getEnv("RABBITMQ_USER", "guest"),
		getEnv("RABBITMQ_PASSWORD", "guest"),
	)
	if err != nil {
		log.Fatal("Failed to connect RabbitMQ consumer: ", err)
	}
	defer consumer.Close()

	// Wire layers
	orderRepo := repository.NewOrderRepository(db)
	orderService := service.NewOrderService(orderRepo, publisher)
	orderHandler := handler.NewOrderHandler(orderService)

	// Start consuming inventory.updated events
	consumer.ConsumeInventoryUpdated(func(data rabbitmq.InventoryUpdatedData) {
		if data.IsLowStock {
			log.Printf("Low stock alert for product %s: %d remaining", data.ProductID, data.QuantityRemaining)
		}
	})

	// Gin router
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"service": "order-service", "status": "running"})
	})

	orderHandler.RegisterRoutes(r)

	port := getEnv("SERVER_PORT", "8003")
	log.Printf("Order Service starting on port %s", port)
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
