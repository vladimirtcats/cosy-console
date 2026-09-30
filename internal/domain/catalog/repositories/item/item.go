package item

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"cosy-console/internal/domain/catalog"
	"cosy-console/internal/domain/catalog/models"
	"cosy-console/pkg/transaction"
)

type ItemRepo struct {
	db *gorm.DB
}

func New(db *gorm.DB) *ItemRepo {
	return &ItemRepo{
		db: db,
	}
}

func (r *ItemRepo) Get(ctx context.Context, id uuid.UUID) (models.Item, error) {
	var item models.Item

	err := transaction.ExtractDB(ctx, r.db).WithContext(ctx).
		First(&item, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Item{}, catalog.ErrNotFound
	}
	if err != nil {
		return models.Item{}, err
	}

	return item, nil
}

func (r *ItemRepo) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Item, error) {
	items := make([]models.Item, 0, len(ids))

	err := transaction.ExtractDB(ctx, r.db).WithContext(ctx).
		Find(&items, "id IN ?", ids).Error
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (r *ItemRepo) List(ctx context.Context, filter catalog.ItemFilter) ([]models.Item, error) {
	query := transaction.ExtractDB(ctx, r.db).WithContext(ctx).Model(&models.Item{})

	if !filter.IncludeInactive {
		query = query.Where("active = ?", true)
	}

	if filter.SKU != "" {
		query = query.Where("sku ILIKE ?", "%"+filter.SKU+"%")
	}

	items := make([]models.Item, 0)

	if err := query.Order("sku ASC").Find(&items).Error; err != nil {
		return nil, err
	}

	return items, nil
}
