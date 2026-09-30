package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusPending   = "pending"
	StatusPaid      = "paid"
	StatusShipped   = "shipped"
	StatusDelivered = "delivered"
	StatusCancelled = "cancelled"
)

type Order struct {
	ID         uuid.UUID
	UserID     int64
	Status     string
	Note       string
	TotalCents int64
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Lines      []OrderLine `gorm:"foreignKey:OrderID"`
}

func (Order) TableName() string {
	return "orders"
}
