package model

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID          uuid.UUID   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID   `gorm:"type:uuid;not null" json:"user_id"`
	Status      string      `gorm:"type:varchar(50);not null;default:'pending'" json:"status"`
	TotalAmount float64     `gorm:"type:decimal(10,2);not null" json:"total_amount"`
	Items       []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
	CreatedAt   time.Time   `gorm:"default:now()" json:"created_at"`
	UpdatedAt   time.Time   `gorm:"default:now()" json:"updated_at"`
}

func (Order) TableName() string {
	return "order_schema.orders"
}

type OrderItem struct {
	ID        int       `gorm:"primaryKey" json:"id"`
	OrderID   uuid.UUID `gorm:"type:uuid" json:"order_id"`
	ProductID uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`
	Quantity  int       `gorm:"not null" json:"quantity"`
	Price     float64   `gorm:"type:decimal(10,2);not null" json:"price"`
}

func (OrderItem) TableName() string {
	return "order_schema.order_items"
}

type Payment struct {
	ID      uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrderID uuid.UUID  `gorm:"type:uuid" json:"order_id"`
	Method  string     `gorm:"type:varchar(50);not null" json:"method"`
	Status  string     `gorm:"type:varchar(50);not null;default:'pending'" json:"status"`
	PaidAt  *time.Time `json:"paid_at"`
}

func (Payment) TableName() string {
	return "order_schema.payments"
}

type OrderItemInput struct {
	ProductID string  `json:"product_id" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required,min=1"`
	Price     float64 `json:"price" binding:"required,gt=0"`
}

type PlaceOrderInput struct {
	UserID string           `json:"user_id"`
	Items  []OrderItemInput `json:"items" binding:"required,min=1"`
}
