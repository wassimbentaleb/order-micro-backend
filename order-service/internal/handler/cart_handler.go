package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hero/microservice/order-service/internal/service"
)

type CartHandler struct {
	cartService service.CartService
}

func NewCartHandler(cartService service.CartService) *CartHandler {
	return &CartHandler{cartService: cartService}
}

func (h *CartHandler) GetCart(c *gin.Context) {
	userID := c.Param("userId")
	items, err := h.cartService.GetCart(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"cart": items})
}

func (h *CartHandler) AddToCart(c *gin.Context) {
	userID := c.Param("userId")
	var item service.CartItem
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	items, err := h.cartService.AddToCart(userID, item)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"cart": items})
}

func (h *CartHandler) RemoveFromCart(c *gin.Context) {
	userID := c.Param("userId")
	productID := c.Param("productId")

	items, err := h.cartService.RemoveFromCart(userID, productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"cart": items})
}

func (h *CartHandler) ClearCart(c *gin.Context) {
	userID := c.Param("userId")
	if err := h.cartService.ClearCart(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "cart cleared"})
}

func (h *CartHandler) RegisterRoutes(r *gin.Engine) {
	cart := r.Group("/api/cart")
	{
		cart.GET("/:userId", h.GetCart)
		cart.POST("/:userId", h.AddToCart)
		cart.DELETE("/:userId/:productId", h.RemoveFromCart)
		cart.DELETE("/:userId", h.ClearCart)
	}
}
