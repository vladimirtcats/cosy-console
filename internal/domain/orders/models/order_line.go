package models

import "github.com/google/uuid"

type OrderLine struct {
	ID         int64
	OrderID    uuid.UUID
	ItemID     uuid.UUID
	Qty        int
	PriceCents int64
}

func (OrderLine) TableName() string {
	return "order_lines"
}
