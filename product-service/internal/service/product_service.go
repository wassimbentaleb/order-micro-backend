package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hero/microservice/product-service/internal/model"
	"github.com/hero/microservice/product-service/internal/rabbitmq"
	"github.com/hero/microservice/product-service/internal/repository"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const productCacheTTL = 10 * time.Minute

type ProductService interface {
	CreateProduct(input model.CreateProductInput) (*model.Product, error)
	GetAllProducts() ([]model.Product, error)
	GetProduct(id uuid.UUID) (*model.Product, error)
	UpdateProduct(id uuid.UUID, input model.UpdateProductInput) (*model.Product, error)
	UpdateStock(id uuid.UUID, input model.UpdateStockInput) (*model.Inventory, error)
	DecrementStock(productID uuid.UUID, amount int) error
}

type productService struct {
	repo      repository.ProductRepository
	publisher *rabbitmq.Publisher
	rdb       *redis.Client
}

func NewProductService(repo repository.ProductRepository, publisher *rabbitmq.Publisher, rdb *redis.Client) ProductService {
	return &productService{repo: repo, publisher: publisher, rdb: rdb}
}

func (s *productService) cacheKey(id uuid.UUID) string {
	return fmt.Sprintf("product:%s", id.String())
}

func (s *productService) cacheProduct(product *model.Product) {
	data, err := json.Marshal(product)
	if err != nil {
		return
	}
	s.rdb.Set(context.Background(), s.cacheKey(product.ID), data, productCacheTTL)
}

func (s *productService) getCachedProduct(id uuid.UUID) *model.Product {
	val, err := s.rdb.Get(context.Background(), s.cacheKey(id)).Result()
	if err != nil {
		return nil
	}
	var product model.Product
	if err := json.Unmarshal([]byte(val), &product); err != nil {
		return nil
	}
	return &product
}

func (s *productService) invalidateCache(id uuid.UUID) {
	s.rdb.Del(context.Background(), s.cacheKey(id))
}

func (s *productService) CreateProduct(input model.CreateProductInput) (*model.Product, error) {
	product := &model.Product{
		ID:          uuid.New(),
		Name:        input.Name,
		Description: input.Description,
		Price:       input.Price,
		CategoryID:  input.CategoryID,
	}

	if err := s.repo.Create(product); err != nil {
		return nil, errors.New("failed to create product: " + err.Error())
	}

	inv := &model.Inventory{
		ProductID: product.ID,
		Quantity:  input.Quantity,
	}
	if err := s.repo.CreateInventory(inv); err != nil {
		return nil, errors.New("failed to create inventory: " + err.Error())
	}

	s.cacheProduct(product)

	s.publisher.Publish("product.created", map[string]interface{}{
		"product_id":   product.ID.String(),
		"product_name": product.Name,
		"price":        product.Price,
	})

	return product, nil
}

func (s *productService) GetAllProducts() ([]model.Product, error) {
	return s.repo.GetAll()
}

func (s *productService) GetProduct(id uuid.UUID) (*model.Product, error) {
	// Try cache first
	if cached := s.getCachedProduct(id); cached != nil {
		return cached, nil
	}

	product, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("product not found")
		}
		return nil, err
	}

	s.cacheProduct(product)
	return product, nil
}

func (s *productService) UpdateProduct(id uuid.UUID, input model.UpdateProductInput) (*model.Product, error) {
	product, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("product not found")
		}
		return nil, err
	}

	if input.Name != "" {
		product.Name = input.Name
	}
	if input.Description != "" {
		product.Description = input.Description
	}
	if input.Price > 0 {
		product.Price = input.Price
	}
	if input.CategoryID != nil {
		product.CategoryID = input.CategoryID
	}

	if err := s.repo.Update(product); err != nil {
		return nil, errors.New("failed to update product: " + err.Error())
	}

	s.invalidateCache(id)
	s.cacheProduct(product)

	return product, nil
}

func (s *productService) UpdateStock(id uuid.UUID, input model.UpdateStockInput) (*model.Inventory, error) {
	if err := s.repo.UpdateStock(id, input.Quantity); err != nil {
		return nil, errors.New("failed to update stock: " + err.Error())
	}

	inv, err := s.repo.GetStock(id)
	if err != nil {
		return nil, err
	}

	s.invalidateCache(id)

	s.publisher.Publish("inventory.updated", map[string]interface{}{
		"product_id":         id.String(),
		"quantity_remaining": inv.Quantity,
		"is_low_stock":       inv.Quantity < 10,
	})

	return inv, nil
}

func (s *productService) DecrementStock(productID uuid.UUID, amount int) error {
	inv, err := s.repo.DecrementStock(productID, amount)
	if err != nil {
		return err
	}

	s.invalidateCache(productID)

	s.publisher.Publish("inventory.updated", map[string]interface{}{
		"product_id":         productID.String(),
		"quantity_remaining": inv.Quantity,
		"is_low_stock":       inv.Quantity < 10,
	})

	if inv.Quantity == 0 {
		product, _ := s.repo.GetByID(productID)
		name := productID.String()
		if product != nil {
			name = product.Name
		}
		s.publisher.Publish("product.out_of_stock", map[string]interface{}{
			"product_id":   productID.String(),
			"product_name": name,
		})
	}

	return nil
}
