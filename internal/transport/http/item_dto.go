package httptransport

import (
	"time"

	"github.com/google/uuid"

	catalogmodels "cosy-console/internal/domain/catalog/models"
)

type ItemResponse struct {
	ID         uuid.UUID `json:"id"`
	SKU        string    `json:"sku"`
	Title      string    `json:"title"`
	PriceCents int64     `json:"price_cents"`
	Active     bool      `json:"active"`
	CreatedAt  time.Time `json:"created_at"`
}

func toItemResponse(item catalogmodels.Item) ItemResponse {
	return ItemResponse{
		ID:         item.ID,
		SKU:        item.SKU,
		Title:      item.Title,
		PriceCents: item.PriceCents,
		Active:     item.Active,
		CreatedAt:  item.CreatedAt.UTC(),
	}
}

func toItemResponses(items []catalogmodels.Item) []ItemResponse {
	result := make([]ItemResponse, 0, len(items))

	for _, item := range items {
		result = append(result, toItemResponse(item))
	}

	return result
}
