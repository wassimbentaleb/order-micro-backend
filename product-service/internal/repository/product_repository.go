package repository

import (
	"github.com/google/uuid"
	"github.com/hero/microservice/product-service/internal/model"
	"gorm.io/gorm"
)

type ProductRepository interface {
	Create(product *model.Product) error
	GetAll() ([]model.Product, error)
	GetByID(id uuid.UUID) (*model.Product, error)
	Update(product *model.Product) error
	CreateInventory(inv *model.Inventory) error
	GetStock(productID uuid.UUID) (*model.Inventory, error)
	UpdateStock(productID uuid.UUID, quantity int) error
	DecrementStock(productID uuid.UUID, amount int) (*model.Inventory, error)
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(product *model.Product) error {
	return r.db.Create(product).Error
}

func (r *productRepository) GetAll() ([]model.Product, error) {
	var products []model.Product
	err := r.db.Preload("Category").Find(&products).Error
	return products, err
}

func (r *productRepository) GetByID(id uuid.UUID) (*model.Product, error) {
	var product model.Product
	err := r.db.Preload("Category").Where("id = ?", id).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) Update(product *model.Product) error {
	return r.db.Save(product).Error
}

func (r *productRepository) CreateInventory(inv *model.Inventory) error {
	return r.db.Create(inv).Error
}

func (r *productRepository) GetStock(productID uuid.UUID) (*model.Inventory, error) {
	var inv model.Inventory
	err := r.db.Where("product_id = ?", productID).First(&inv).Error
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *productRepository) UpdateStock(productID uuid.UUID, quantity int) error {
	return r.db.Model(&model.Inventory{}).Where("product_id = ?", productID).
		Update("quantity", quantity).Error
}

func (r *productRepository) DecrementStock(productID uuid.UUID, amount int) (*model.Inventory, error) {
	var inv model.Inventory
	err := r.db.Where("product_id = ?", productID).First(&inv).Error
	if err != nil {
		return nil, err
	}

	inv.Quantity -= amount
	if inv.Quantity < 0 {
		inv.Quantity = 0
	}

	err = r.db.Save(&inv).Error
	return &inv, err
}
