package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const cartTTL = 7 * 24 * time.Hour // 7 days

type CartItem struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type CartService interface {
	GetCart(userID string) ([]CartItem, error)
	AddToCart(userID string, item CartItem) ([]CartItem, error)
	RemoveFromCart(userID string, productID string) ([]CartItem, error)
	ClearCart(userID string) error
}

type cartService struct {
	rdb *redis.Client
}

func NewCartService(rdb *redis.Client) CartService {
	return &cartService{rdb: rdb}
}

func (s *cartService) cartKey(userID string) string {
	return fmt.Sprintf("cart:%s", userID)
}

func (s *cartService) GetCart(userID string) ([]CartItem, error) {
	val, err := s.rdb.Get(context.Background(), s.cartKey(userID)).Result()
	if err == redis.Nil {
		return []CartItem{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get cart: %w", err)
	}

	var items []CartItem
	if err := json.Unmarshal([]byte(val), &items); err != nil {
		return nil, fmt.Errorf("failed to parse cart: %w", err)
	}
	return items, nil
}

func (s *cartService) AddToCart(userID string, item CartItem) ([]CartItem, error) {
	items, err := s.GetCart(userID)
	if err != nil {
		return nil, err
	}

	// Update quantity if product already in cart
	found := false
	for i, existing := range items {
		if existing.ProductID == item.ProductID {
			items[i].Quantity += item.Quantity
			items[i].Price = item.Price
			found = true
			break
		}
	}
	if !found {
		items = append(items, item)
	}

	data, _ := json.Marshal(items)
	s.rdb.Set(context.Background(), s.cartKey(userID), data, cartTTL)

	return items, nil
}

func (s *cartService) RemoveFromCart(userID string, productID string) ([]CartItem, error) {
	items, err := s.GetCart(userID)
	if err != nil {
		return nil, err
	}

	var updated []CartItem
	for _, item := range items {
		if item.ProductID != productID {
			updated = append(updated, item)
		}
	}

	if len(updated) == 0 {
		s.rdb.Del(context.Background(), s.cartKey(userID))
		return []CartItem{}, nil
	}

	data, _ := json.Marshal(updated)
	s.rdb.Set(context.Background(), s.cartKey(userID), data, cartTTL)

	return updated, nil
}

func (s *cartService) ClearCart(userID string) error {
	return s.rdb.Del(context.Background(), s.cartKey(userID)).Err()
}
