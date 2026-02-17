package model

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID   int    `gorm:"primaryKey" json:"id"`
	Name string `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
}

func (Category) TableName() string {
	return "product_schema.categories"
}

type Product struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"type:varchar(200);not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Price       float64   `gorm:"type:decimal(10,2);not null" json:"price"`
	CategoryID  *int      `gorm:"index" json:"category_id"`
	Category    *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	CreatedAt   time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt   time.Time `gorm:"default:now()" json:"updated_at"`
}

func (Product) TableName() string {
	return "product_schema.products"
}

type Inventory struct {
	ID        int       `gorm:"primaryKey" json:"id"`
	ProductID uuid.UUID `gorm:"type:uuid;uniqueIndex" json:"product_id"`
	Quantity  int       `gorm:"not null;default:0" json:"quantity"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`
}

func (Inventory) TableName() string {
	return "product_schema.inventory"
}

type CreateProductInput struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"required,gt=0"`
	CategoryID  *int    `json:"category_id"`
	Quantity    int     `json:"quantity" binding:"min=0"`
}

type UpdateProductInput struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"omitempty,gt=0"`
	CategoryID  *int    `json:"category_id"`
}

type UpdateStockInput struct {
	Quantity int `json:"quantity" binding:"required,min=0"`
}
