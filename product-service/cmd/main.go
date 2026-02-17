package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hero/microservice/product-service/internal/handler"
	"github.com/hero/microservice/product-service/internal/rabbitmq"
	"github.com/hero/microservice/product-service/internal/repository"
	"github.com/hero/microservice/product-service/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s search_path=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "svc_product"),
		getEnv("DB_PASSWORD", "svc_product_pass"),
		getEnv("DB_NAME", "microservice_db"),
		getEnv("DB_SCHEMA", "product_schema"),
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
	productRepo := repository.NewProductRepository(db)
	productService := service.NewProductService(productRepo, publisher)
	productHandler := handler.NewProductHandler(productService)

	// Start consuming order.created events
	consumer.ConsumeOrderCreated(func(productID uuid.UUID, quantity int) error {
		return productService.DecrementStock(productID, quantity)
	})

	// Gin router
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"service": "product-service", "status": "running"})
	})

	productHandler.RegisterRoutes(r)

	port := getEnv("SERVER_PORT", "8002")
	log.Printf("Product Service starting on port %s", port)
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
