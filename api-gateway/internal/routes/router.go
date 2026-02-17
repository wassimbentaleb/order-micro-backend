package routes

import (
	"log"
	"net/http"

	"github.com/hero/microservice/api-gateway/internal/proxy"
)

type ServiceConfig struct {
	UserServiceURL         string
	ProductServiceURL      string
	OrderServiceURL        string
	NotificationServiceURL string
}

func SetupRoutes(mux *http.ServeMux, cfg ServiceConfig) {
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

	mux.Handle("/api/users/", userProxy)
	mux.Handle("/api/products/", productProxy)
	mux.Handle("/api/products", productProxy)
	mux.Handle("/api/orders/", orderProxy)
	mux.Handle("/api/orders", orderProxy)
	mux.Handle("/api/notifications/", notifProxy)
}
