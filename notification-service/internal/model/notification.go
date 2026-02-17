package model

import (
	"time"

	"github.com/google/uuid"
)

type Template struct {
	ID           int    `gorm:"primaryKey" json:"id"`
	Name         string `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	BodyTemplate string `gorm:"type:text;not null" json:"body_template"`
	Type         string `gorm:"type:varchar(20);not null" json:"type"`
}

func (Template) TableName() string {
	return "notification_schema.templates"
}

type NotifLog struct {
	ID      uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID  uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	Type    string    `gorm:"type:varchar(20);not null" json:"type"`
	Subject string    `gorm:"type:varchar(255)" json:"subject"`
	Body    string    `gorm:"type:text" json:"body"`
	Status  string    `gorm:"type:varchar(20);default:'sent'" json:"status"`
	SentAt  time.Time `gorm:"default:now()" json:"sent_at"`
}

func (NotifLog) TableName() string {
	return "notification_schema.notif_logs"
}
