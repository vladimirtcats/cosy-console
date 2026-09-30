package models

import (
	"time"

	"github.com/google/uuid"
)

type Item struct {
	ID         uuid.UUID
	SKU        string
	Title      string
	PriceCents int64
	Active     bool
	CreatedAt  time.Time
}

func (Item) TableName() string {
	return "items"
}
