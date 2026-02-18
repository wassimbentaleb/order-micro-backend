package routes

import (
	"log"
	"net/http"

	"github.com/hero/microservice/api-gateway/internal/middleware"
	"github.com/hero/microservice/api-gateway/internal/proxy"
	"github.com/redis/go-redis/v9"
)

type ServiceConfig struct {
	UserServiceURL         string
	ProductServiceURL      string
	OrderServiceURL        string
	NotificationServiceURL string
}

func SetupRoutes(mux *http.ServeMux, cfg ServiceConfig, rdb *redis.Client) {
	userProxy, err := proxy.NewReverseProxy(cfg.UserServiceURL)
	if err != nil {
		log.Fatal("Failed to create user service proxy: ", err)
	}

	productProxy, err := proxy.NewReverseProxy(cfg.ProductServiceURL)
	if err != nil {
		log.Fatal("Failed to create product service proxy: ", err)
	}

	orderProxy, err := proxy.NewReverseProxy(cfg.OrderServiceURL)
	if err != nil {
		log.Fatal("Failed to create order service proxy: ", err)
	}

	notifProxy, err := proxy.NewReverseProxy(cfg.NotificationServiceURL)
	if err != nil {
		log.Fatal("Failed to create notification service proxy: ", err)
	}

	auth := middleware.AuthMiddleware(rdb)

	// Public routes (no auth)
	mux.Handle("/api/users/register", userProxy)
	mux.Handle("/api/users/login", userProxy)

	// Protected routes (require auth)
	mux.Handle("/api/users/logout", auth(userProxy))
	mux.Handle("/api/users/me", auth(userProxy))
	mux.Handle("/api/users/", auth(userProxy))

	mux.Handle("/api/products/", auth(productProxy))
	mux.Handle("/api/products", auth(productProxy))

	mux.Handle("/api/orders/", auth(orderProxy))
	mux.Handle("/api/orders", auth(orderProxy))

	mux.Handle("/api/cart/", auth(orderProxy))
	mux.Handle("/api/cart", auth(orderProxy))

	mux.Handle("/api/notifications/", auth(notifProxy))
}
