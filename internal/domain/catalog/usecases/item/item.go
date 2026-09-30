package item

import (
	"context"

	"github.com/google/uuid"

	"cosy-console/internal/domain/catalog"
	"cosy-console/internal/domain/catalog/models"
	"cosy-console/pkg/pagination"
)

type itemRepo interface {
	Get(ctx context.Context, id uuid.UUID) (models.Item, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Item, error)
	List(ctx context.Context, filter catalog.ItemFilter) ([]models.Item, error)
}

// ActiveItemsCache is a best-effort cache: Fetch reports a miss without an
// error, and Store never fails the caller — a broken cache must not break
// the listing.
type ActiveItemsCache interface {
	Fetch(ctx context.Context) (items []models.Item, found bool)
	Store(ctx context.Context, items []models.Item)
}

type ItemUseCase struct {
	itemRepo itemRepo
	cache    ActiveItemsCache
}

func New(itemRepo itemRepo, cache ActiveItemsCache) *ItemUseCase {
	return &ItemUseCase{
		itemRepo: itemRepo,
		cache:    cache,
	}
}

func (uc *ItemUseCase) Get(ctx context.Context, id uuid.UUID) (models.Item, error) {
	return uc.itemRepo.Get(ctx, id)
}

func (uc *ItemUseCase) List(
	ctx context.Context,
	filter catalog.ItemFilter,
	opts ...pagination.Option,
) ([]models.Item, pagination.Result, error) {
	settings := pagination.NewSettings(opts...)

	items, err := uc.loadItems(ctx, filter)
	if err != nil {
		return nil, pagination.Result{}, err
	}

	page := slicePage(items, settings)

	return page, pagination.NewResult(int64(len(items)), settings), nil
}

// loadItems serves the default listing from the cache; any filtered listing
// goes straight to the repository.
func (uc *ItemUseCase) loadItems(ctx context.Context, filter catalog.ItemFilter) ([]models.Item, error) {
	if !filter.IsDefault() {
		return uc.itemRepo.List(ctx, filter)
	}

	if cached, found := uc.cache.Fetch(ctx); found {
		return cached, nil
	}

	loaded, err := uc.itemRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	uc.cache.Store(ctx, loaded)

	return loaded, nil
}

func slicePage(items []models.Item, settings pagination.Settings) []models.Item {
	start := settings.Offset()
	if start >= len(items) {
		return []models.Item{}
	}

	end := start + settings.PageSize
	if end > len(items) {
		end = len(items)
	}

	return items[start:end]
}
