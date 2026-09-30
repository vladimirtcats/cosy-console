package adapters

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"cosy-console/internal/domain/catalog"
	"cosy-console/internal/domain/catalog/models"
	"cosy-console/internal/domain/orders"
)

type itemsReader interface {
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Item, error)
}

// CatalogPricer adapts the catalog domain to the orders.ItemPricer port:
// it prices order lines with current catalog prices.
type CatalogPricer struct {
	items itemsReader
}

func NewCatalogPricer(items itemsReader) *CatalogPricer {
	return &CatalogPricer{
		items: items,
	}
}

func (p *CatalogPricer) PriceItems(ctx context.Context, lines []orders.LineInput) ([]orders.PricedLine, error) {
	items, err := p.loadItems(ctx, lines)
	if err != nil {
		return nil, err
	}

	byID := make(map[uuid.UUID]models.Item, len(items))
	for _, item := range items {
		byID[item.ID] = item
	}

	priced := make([]orders.PricedLine, 0, len(lines))
	for _, line := range lines {
		item, found := byID[line.ItemID]
		if !found || !item.Active {
			return nil, fmt.Errorf("%w: %s", catalog.ErrNotFound, line.ItemID)
		}

		priced = append(priced, orders.PricedLine{
			ItemID:     item.ID,
			Qty:        line.Qty,
			PriceCents: item.PriceCents,
		})
	}

	return priced, nil
}

func (p *CatalogPricer) loadItems(ctx context.Context, lines []orders.LineInput) ([]models.Item, error) {
	ids := make([]uuid.UUID, 0, len(lines))
	for _, line := range lines {
		ids = append(ids, line.ItemID)
	}

	items, err := p.items.GetByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("load items: %w", err)
	}

	return items, nil
}
