package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/hero/microservice/notification-service/internal/handler"
	"github.com/hero/microservice/notification-service/internal/rabbitmq"
	"github.com/hero/microservice/notification-service/internal/repository"
	"github.com/hero/microservice/notification-service/internal/service"
	"github.com/hero/microservice/pkg/cache"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s search_path=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "svc_notif"),
		getEnv("DB_PASSWORD", "svc_notif_pass"),
		getEnv("DB_NAME", "microservice_db"),
		getEnv("DB_SCHEMA", "notification_schema"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}
	log.Println("Database connected")

	// Redis
	rdb, err := cache.NewRedisClient(
		getEnv("REDIS_HOST", "localhost"),
		getEnv("REDIS_PORT", "6379"),
	)
	if err != nil {
		log.Fatal("Failed to connect to Redis: ", err)
	}
	defer rdb.Close()

	// RabbitMQ consumer (with Redis for deduplication)
	consumer, err := rabbitmq.NewConsumer(
		getEnv("RABBITMQ_HOST", "localhost"),
		getEnv("RABBITMQ_PORT", "5672"),
		getEnv("RABBITMQ_USER", "guest"),
		getEnv("RABBITMQ_PASSWORD", "guest"),
		rdb,
	)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ: ", err)
	}
	defer consumer.Close()

	// Wire layers
	notifRepo := repository.NewNotificationRepository(db)
	notifService := service.NewNotificationService(notifRepo)
	notifHandler := handler.NewNotificationHandler(notifService)

	// Start all consumers
	consumer.StartConsuming(rabbitmq.EventHandlers{
		OnUserRegistered:    notifService.HandleUserRegistered,
		OnOrderCreated:      notifService.HandleOrderCreated,
		OnOrderCompleted:    notifService.HandleOrderCompleted,
		OnProductOutOfStock: notifService.HandleProductOutOfStock,
	})

	// Gin router
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"service": "notification-service", "status": "running"})
	})

	notifHandler.RegisterRoutes(r)

	port := getEnv("SERVER_PORT", "8004")
	log.Printf("Notification Service starting on port %s", port)
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
