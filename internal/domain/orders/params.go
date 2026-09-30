package orders

import "github.com/google/uuid"

type CreationParams struct {
	UserID int64
	Lines  []LineInput
}

type LineInput struct {
	ItemID uuid.UUID
	Qty    int
}

type PricedLine struct {
	ItemID     uuid.UUID
	Qty        int
	PriceCents int64
}

type UpdateAttrs struct {
	Status *string
	Note   *string
}

type OrderFilter struct {
	Status string
}
