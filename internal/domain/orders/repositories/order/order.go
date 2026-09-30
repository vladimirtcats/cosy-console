package order

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"cosy-console/internal/domain/orders"
	"cosy-console/internal/domain/orders/models"
	"cosy-console/pkg/pagination"
	"cosy-console/pkg/transaction"
)

type OrderRepo struct {
	db *gorm.DB
}

func New(db *gorm.DB) *OrderRepo {
	return &OrderRepo{db: db}
}

func (r *OrderRepo) Get(ctx context.Context, id uuid.UUID) (models.Order, error) {
	var order models.Order
	err := transaction.ExtractDB(ctx, r.db).WithContext(ctx).
		Preload("Lines").
		First(&order, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Order{}, orders.ErrNotFound
	}
	if err != nil {
		return models.Order{}, err
	}

	return order, nil
}

func (r *OrderRepo) ListByUser(
	ctx context.Context,
	userID int64,
	filter orders.OrderFilter,
	opts ...pagination.Option,
) ([]models.Order, pagination.Result, error) {
	settings := pagination.NewSettings(opts...)
	query := transaction.ExtractDB(ctx, r.db).WithContext(ctx).
		Model(&models.Order{}).
		Where("user_id = ?", userID)
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, pagination.Result{}, err
	}

	var ordersList []models.Order
	err := query.
		Preload("Lines").
		// id as a tie-breaker: rows with equal created_at (seeds, tests)
		// must keep a stable order across pages.
		Order("created_at DESC, id DESC").
		Limit(settings.PageSize).
		Offset(settings.Offset()).
		Find(&ordersList).Error
	if err != nil {
		return nil, pagination.Result{}, err
	}

	return ordersList, pagination.NewResult(total, settings), nil
}

func (r *OrderRepo) Create(ctx context.Context, order models.Order) error {
	return transaction.ExtractDB(ctx, r.db).WithContext(ctx).Create(&order).Error
}

func (r *OrderRepo) Update(ctx context.Context, id uuid.UUID, attrs orders.UpdateAttrs) (models.Order, error) {
	if attrs.Status == nil && attrs.Note == nil {
		return r.Get(ctx, id)
	}

	err := transaction.ExtractDB(ctx, r.db).WithContext(ctx).
		Model(&models.Order{}).
		Where("id = ?", id).
		Updates(attrs).Error
	if err != nil {
		return models.Order{}, err
	}

	return r.Get(ctx, id)
}
