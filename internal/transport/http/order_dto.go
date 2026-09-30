package httptransport

import (
	"time"

	"github.com/google/uuid"

	ordersmodels "cosy-console/internal/domain/orders/models"
)

type OrderLineResponse struct {
	ID         int64     `json:"id"`
	ItemID     uuid.UUID `json:"item_id"`
	Qty        int       `json:"qty"`
	PriceCents int64     `json:"price_cents"`
}

type OrderResponse struct {
	ID         uuid.UUID           `json:"id"`
	UserID     int64               `json:"user_id"`
	Status     string              `json:"status"`
	Note       string              `json:"note"`
	TotalCents int64               `json:"total_cents"`
	CreatedAt  time.Time           `json:"created_at"`
	Lines      []OrderLineResponse `json:"lines"`
}

func toOrderResponse(order ordersmodels.Order) OrderResponse {
	lines := make([]OrderLineResponse, 0, len(order.Lines))

	for _, line := range order.Lines {
		lines = append(lines, OrderLineResponse{
			ID:         line.ID,
			ItemID:     line.ItemID,
			Qty:        line.Qty,
			PriceCents: line.PriceCents,
		})
	}

	return OrderResponse{
		ID:         order.ID,
		UserID:     order.UserID,
		Status:     order.Status,
		Note:       order.Note,
		TotalCents: order.TotalCents,
		CreatedAt:  order.CreatedAt.UTC(),
		Lines:      lines,
	}
}

func toOrderResponses(orders []ordersmodels.Order) []OrderResponse {
	result := make([]OrderResponse, 0, len(orders))

	for _, order := range orders {
		result = append(result, toOrderResponse(order))
	}

	return result
}
