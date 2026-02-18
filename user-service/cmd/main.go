package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/hero/microservice/pkg/cache"
	"github.com/hero/microservice/user-service/internal/handler"
	"github.com/hero/microservice/user-service/internal/rabbitmq"
	"github.com/hero/microservice/user-service/internal/repository"
	"github.com/hero/microservice/user-service/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Database connection
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s search_path=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "svc_user"),
		getEnv("DB_PASSWORD", "svc_user_pass"),
		getEnv("DB_NAME", "microservice_db"),
		getEnv("DB_SCHEMA", "user_schema"),
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

	// Redis
	rdb, err := cache.NewRedisClient(
		getEnv("REDIS_HOST", "localhost"),
		getEnv("REDIS_PORT", "6379"),
	)
	if err != nil {
		log.Fatal("Failed to connect to Redis: ", err)
	}
	defer rdb.Close()

	// Wire layers
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo, publisher, rdb)
	userHandler := handler.NewUserHandler(userService)

	// Gin router
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"service": "user-service", "status": "running"})
	})

	userHandler.RegisterRoutes(r)

	port := getEnv("SERVER_PORT", "8001")
	log.Printf("User Service starting on port %s", port)
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
