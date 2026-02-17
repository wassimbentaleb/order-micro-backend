package repository

import (
	"github.com/google/uuid"
	"github.com/hero/microservice/order-service/internal/model"
	"gorm.io/gorm"
)

type OrderRepository interface {
	Create(order *model.Order) error
	GetByID(id uuid.UUID) (*model.Order, error)
	GetByUserID(userID uuid.UUID) ([]model.Order, error)
	UpdateStatus(id uuid.UUID, status string) error
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(order *model.Order) error {
	return r.db.Create(order).Error
}

func (r *orderRepository) GetByID(id uuid.UUID) (*model.Order, error) {
	var order model.Order
	err := r.db.Preload("Items").Where("id = ?", id).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) GetByUserID(userID uuid.UUID) ([]model.Order, error) {
	var orders []model.Order
	err := r.db.Preload("Items").Where("user_id = ?", userID).
		Order("created_at DESC").Find(&orders).Error
	return orders, err
}

func (r *orderRepository) UpdateStatus(id uuid.UUID, status string) error {
	return r.db.Model(&model.Order{}).Where("id = ?", id).Update("status", status).Error
}
