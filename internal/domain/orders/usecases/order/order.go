package order

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"cosy-console/internal/domain/orders"
	"cosy-console/internal/domain/orders/models"
	"cosy-console/pkg/pagination"
	"cosy-console/pkg/transaction"
)

type orderRepo interface {
	Get(ctx context.Context, id uuid.UUID) (models.Order, error)
	ListByUser(ctx context.Context, userID int64, filter orders.OrderFilter, opts ...pagination.Option) ([]models.Order, pagination.Result, error)
	Create(ctx context.Context, order models.Order) error
	Update(ctx context.Context, id uuid.UUID, attrs orders.UpdateAttrs) (models.Order, error)
}

// ItemPricer is the catalog port: the orders domain prices new lines
// through it without importing the catalog domain.
type ItemPricer interface {
	PriceItems(ctx context.Context, lines []orders.LineInput) ([]orders.PricedLine, error)
}

type OrderUseCase struct {
	orderRepo  orderRepo
	itemPricer ItemPricer
	tx         transaction.Manager
}

func New(orderRepo orderRepo, itemPricer ItemPricer, tx transaction.Manager) *OrderUseCase {
	return &OrderUseCase{
		orderRepo:  orderRepo,
		itemPricer: itemPricer,
		tx:         tx,
	}
}

func (uc *OrderUseCase) Get(ctx context.Context, id uuid.UUID) (models.Order, error) {
	return uc.orderRepo.Get(ctx, id)
}

func (uc *OrderUseCase) ListByUser(
	ctx context.Context,
	userID int64,
	filter orders.OrderFilter,
	opts ...pagination.Option,
) ([]models.Order, pagination.Result, error) {
	return uc.orderRepo.ListByUser(ctx, userID, filter, opts...)
}

func (uc *OrderUseCase) Create(ctx context.Context, params orders.CreationParams) (models.Order, error) {
	if params.UserID == 0 {
		return models.Order{}, errors.New("user_id is required")
	}
	if len(params.Lines) == 0 {
		return models.Order{}, errors.New("order must contain at least one line")
	}

	priced, err := uc.itemPricer.PriceItems(ctx, params.Lines)
	if err != nil {
		return models.Order{}, fmt.Errorf("price items: %w", err)
	}

	order := buildOrder(params.UserID, priced)

	if err := uc.tx.InTransaction(ctx, func(ctx context.Context) error {
		return uc.orderRepo.Create(ctx, order)
	}); err != nil {
		return models.Order{}, fmt.Errorf("create order: %w", err)
	}

	// Re-read: the persisted row carries timestamps and other DB defaults.
	return uc.orderRepo.Get(ctx, order.ID)
}

func (uc *OrderUseCase) Update(ctx context.Context, id uuid.UUID, attrs orders.UpdateAttrs) (models.Order, error) {
	return uc.orderRepo.Update(ctx, id, attrs)
}

func (uc *OrderUseCase) Cancel(ctx context.Context, id uuid.UUID) (models.Order, error) {
	order, err := uc.orderRepo.Get(ctx, id)
	if err != nil {
		return models.Order{}, err
	}

	if order.Status != models.StatusPending && order.Status != models.StatusPaid {
		return models.Order{}, fmt.Errorf("%w: %s -> %s", orders.ErrInvalidTransition, order.Status, models.StatusCancelled)
	}

	cancelled := models.StatusCancelled

	return uc.orderRepo.Update(ctx, id, orders.UpdateAttrs{Status: &cancelled})
}

func buildOrder(userID int64, priced []orders.PricedLine) models.Order {
	order := models.Order{
		ID:     uuid.New(),
		UserID: userID,
		Status: models.StatusPending,
		Lines:  make([]models.OrderLine, 0, len(priced)),
	}

	for _, line := range priced {
		order.TotalCents += line.PriceCents * int64(line.Qty)
		order.Lines = append(order.Lines, models.OrderLine{
			OrderID:    order.ID,
			ItemID:     line.ItemID,
			Qty:        line.Qty,
			PriceCents: line.PriceCents,
		})
	}

	return order
}
