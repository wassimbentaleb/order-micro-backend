package service

import (
	"errors"

	"github.com/google/uuid"
	"github.com/hero/microservice/order-service/internal/model"
	"github.com/hero/microservice/order-service/internal/rabbitmq"
	"github.com/hero/microservice/order-service/internal/repository"
	"gorm.io/gorm"
)

type OrderService interface {
	PlaceOrder(input model.PlaceOrderInput) (*model.Order, error)
	GetOrder(id uuid.UUID) (*model.Order, error)
	GetUserOrders(userID uuid.UUID) ([]model.Order, error)
	CancelOrder(id uuid.UUID) error
}

type orderService struct {
	repo      repository.OrderRepository
	publisher *rabbitmq.Publisher
}

func NewOrderService(repo repository.OrderRepository, publisher *rabbitmq.Publisher) OrderService {
	return &orderService{repo: repo, publisher: publisher}
}

func (s *orderService) PlaceOrder(input model.PlaceOrderInput) (*model.Order, error) {
	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		return nil, errors.New("invalid user_id")
	}

	var totalAmount float64
	var items []model.OrderItem

	for _, item := range input.Items {
		productID, err := uuid.Parse(item.ProductID)
		if err != nil {
			return nil, errors.New("invalid product_id: " + item.ProductID)
		}

		items = append(items, model.OrderItem{
			ProductID: productID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		})

		totalAmount += item.Price * float64(item.Quantity)
	}

	order := &model.Order{
		ID:          uuid.New(),
		UserID:      userID,
		Status:      "pending",
		TotalAmount: totalAmount,
		Items:       items,
	}

	if err := s.repo.Create(order); err != nil {
		return nil, errors.New("failed to create order: " + err.Error())
	}

	// Build event items
	eventItems := make([]map[string]interface{}, len(input.Items))
	for i, item := range input.Items {
		eventItems[i] = map[string]interface{}{
			"product_id": item.ProductID,
			"quantity":   item.Quantity,
		}
	}

	s.publisher.Publish("order.created", map[string]interface{}{
		"order_id":     order.ID.String(),
		"user_id":      order.UserID.String(),
		"items":        eventItems,
		"total_amount": order.TotalAmount,
	})

	return order, nil
}

func (s *orderService) GetOrder(id uuid.UUID) (*model.Order, error) {
	order, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("order not found")
		}
		return nil, err
	}
	return order, nil
}

func (s *orderService) GetUserOrders(userID uuid.UUID) ([]model.Order, error) {
	return s.repo.GetByUserID(userID)
}

func (s *orderService) CancelOrder(id uuid.UUID) error {
	order, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("order not found")
		}
		return err
	}

	if order.Status == "cancelled" {
		return errors.New("order already cancelled")
	}

	if err := s.repo.UpdateStatus(id, "cancelled"); err != nil {
		return errors.New("failed to cancel order: " + err.Error())
	}

	s.publisher.Publish("order.cancelled", map[string]interface{}{
		"order_id": id.String(),
		"user_id":  order.UserID.String(),
	})

	return nil
}
