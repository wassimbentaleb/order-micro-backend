package repository

import (
	"github.com/google/uuid"
	"github.com/hero/microservice/notification-service/internal/model"
	"gorm.io/gorm"
)

type NotificationRepository interface {
	SaveLog(log *model.NotifLog) error
	GetByUserID(userID uuid.UUID) ([]model.NotifLog, error)
	GetTemplate(name string) (*model.Template, error)
}

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) SaveLog(notifLog *model.NotifLog) error {
	return r.db.Create(notifLog).Error
}

func (r *notificationRepository) GetByUserID(userID uuid.UUID) ([]model.NotifLog, error) {
	var logs []model.NotifLog
	err := r.db.Where("user_id = ?", userID).Order("sent_at DESC").Find(&logs).Error
	return logs, err
}

func (r *notificationRepository) GetTemplate(name string) (*model.Template, error) {
	var tmpl model.Template
	err := r.db.Where("name = ?", name).First(&tmpl).Error
	if err != nil {
		return nil, err
	}
	return &tmpl, nil
}
