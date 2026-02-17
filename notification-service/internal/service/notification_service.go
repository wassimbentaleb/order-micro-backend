package service

import (
	"log"

	"github.com/google/uuid"
	"github.com/hero/microservice/notification-service/internal/model"
	"github.com/hero/microservice/notification-service/internal/repository"
)

type NotificationService interface {
	HandleUserRegistered(userID, username, email string)
	HandleOrderCreated(orderID, userID string)
	HandleOrderCompleted(orderID, userID string)
	HandleProductOutOfStock(productID, productName string)
	GetUserNotifications(userID uuid.UUID) ([]model.NotifLog, error)
}

type notificationService struct {
	repo repository.NotificationRepository
}

func NewNotificationService(repo repository.NotificationRepository) NotificationService {
	return &notificationService{repo: repo}
}

func (s *notificationService) HandleUserRegistered(userID, username, email string) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		log.Printf("Invalid user_id in user.registered: %s", userID)
		return
	}

	notifLog := &model.NotifLog{
		ID:      uuid.New(),
		UserID:  uid,
		Type:    "email",
		Subject: "Welcome to our platform!",
		Body:    "Welcome " + username + "! Your account has been created with email " + email,
		Status:  "sent",
	}

	if err := s.repo.SaveLog(notifLog); err != nil {
		log.Printf("Failed to save notification log: %v", err)
	}

	log.Printf("Welcome email sent to user %s (%s)", username, email)
}

func (s *notificationService) HandleOrderCreated(orderID, userID string) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		log.Printf("Invalid user_id in order.created: %s", userID)
		return
	}

	notifLog := &model.NotifLog{
		ID:      uuid.New(),
		UserID:  uid,
		Type:    "email",
		Subject: "Order Confirmation",
		Body:    "Your order #" + orderID + " has been placed successfully.",
		Status:  "sent",
	}

	if err := s.repo.SaveLog(notifLog); err != nil {
		log.Printf("Failed to save notification log: %v", err)
	}

	log.Printf("Order confirmation sent for order %s", orderID)
}

func (s *notificationService) HandleOrderCompleted(orderID, userID string) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		log.Printf("Invalid user_id in order.completed: %s", userID)
		return
	}

	notifLog := &model.NotifLog{
		ID:      uuid.New(),
		UserID:  uid,
		Type:    "email",
		Subject: "Order Delivered",
		Body:    "Your order #" + orderID + " has been delivered.",
		Status:  "sent",
	}

	if err := s.repo.SaveLog(notifLog); err != nil {
		log.Printf("Failed to save notification log: %v", err)
	}

	log.Printf("Order completed notification sent for order %s", orderID)
}

func (s *notificationService) HandleProductOutOfStock(productID, productName string) {
	// Use a placeholder admin UUID for admin notifications
	adminUID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	notifLog := &model.NotifLog{
		ID:      uuid.New(),
		UserID:  adminUID,
		Type:    "email",
		Subject: "Stock Alert: " + productName,
		Body:    "Product " + productName + " (ID: " + productID + ") is out of stock.",
		Status:  "sent",
	}

	if err := s.repo.SaveLog(notifLog); err != nil {
		log.Printf("Failed to save notification log: %v", err)
	}

	log.Printf("Stock alert sent for product %s (%s)", productName, productID)
}

func (s *notificationService) GetUserNotifications(userID uuid.UUID) ([]model.NotifLog, error) {
	return s.repo.GetByUserID(userID)
}
