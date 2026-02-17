package routes

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/hero/microservice/api-gateway/internal/proxy"
)

type ServiceConfig struct {
	UserServiceURL         string
	ProductServiceURL      string
	OrderServiceURL        string
	NotificationServiceURL string
}

func SetupRoutes(r *gin.Engine, cfg ServiceConfig) {
	// User Service proxy
	userProxy, err := proxy.NewReverseProxy(cfg.UserServiceURL)
	if err != nil {
		log.Fatal("Failed to create user service proxy: ", err)
	}

	// Product Service proxy
	productProxy, err := proxy.NewReverseProxy(cfg.ProductServiceURL)
	if err != nil {
		log.Fatal("Failed to create product service proxy: ", err)
	}

	// Order Service proxy
	orderProxy, err := proxy.NewReverseProxy(cfg.OrderServiceURL)
	if err != nil {
		log.Fatal("Failed to create order service proxy: ", err)
	}

	// Notification Service proxy
	notifProxy, err := proxy.NewReverseProxy(cfg.NotificationServiceURL)
	if err != nil {
		log.Fatal("Failed to create notification service proxy: ", err)
	}

	// Route groups
	r.Any("/api/users/*path", gin.WrapH(userProxy))
	r.Any("/api/products/*path", gin.WrapH(productProxy))
	r.Any("/api/orders/*path", gin.WrapH(orderProxy))
	r.Any("/api/notifications/*path", gin.WrapH(notifProxy))
}
