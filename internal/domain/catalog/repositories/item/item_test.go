package item_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"cosy-console/internal/domain/catalog"
	catalogmodels "cosy-console/internal/domain/catalog/models"
	itemrepo "cosy-console/internal/domain/catalog/repositories/item"
	"cosy-console/internal/testdb"
)

var (
	mugID  = uuid.MustParse("00000000-0000-0000-0000-0000000000a1")
	teeID  = uuid.MustParse("00000000-0000-0000-0000-0000000000a3")
	deadID = uuid.MustParse("00000000-0000-0000-0000-0000000000a6")
)

func seedItem(t *testing.T, db *gorm.DB, id uuid.UUID, sku string, active bool) {
	t.Helper()

	err := db.Exec(`
		INSERT INTO items (id, sku, title, price_cents, active)
		VALUES (?, ?, ?, 100, ?)`,
		id, sku, sku, active,
	).Error
	require.NoError(t, err)
}

func TestItemRepo_Get(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, db *gorm.DB) uuid.UUID
		wantErr error
		verify  func(t *testing.T, item catalogmodels.Item)
	}{
		{
			name: "returns the item by id",
			setup: func(t *testing.T, db *gorm.DB) uuid.UUID {
				seedItem(t, db, mugID, "MUG-01", true)

				return mugID
			},
			verify: func(t *testing.T, item catalogmodels.Item) {
				assert.Equal(t, "MUG-01", item.SKU)
				assert.True(t, item.Active)
			},
		},
		{
			name: "missing item maps to the domain ErrNotFound sentinel",
			setup: func(t *testing.T, _ *gorm.DB) uuid.UUID {
				return uuid.New()
			},
			wantErr: catalog.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := testdb.Open(t)
			repo := itemrepo.New(db)

			id := tt.setup(t, db)

			item, err := repo.Get(context.Background(), id)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)

				return
			}

			require.NoError(t, err)
			tt.verify(t, item)
		})
	}
}

func TestItemRepo_GetByIDs(t *testing.T) {
	t.Run("returns the found items regardless of order", func(t *testing.T) {
		db := testdb.Open(t)
		repo := itemrepo.New(db)

		seedItem(t, db, mugID, "MUG-01", true)
		seedItem(t, db, teeID, "TEE-01", true)
		seedItem(t, db, deadID, "RETRO-POSTER", false)

		items, err := repo.GetByIDs(context.Background(), []uuid.UUID{teeID, mugID})

		require.NoError(t, err)
		assert.Len(t, items, 2)
	})

	t.Run("empty id list returns an empty slice", func(t *testing.T) {
		db := testdb.Open(t)
		repo := itemrepo.New(db)

		items, err := repo.GetByIDs(context.Background(), nil)

		require.NoError(t, err)
		require.NotNil(t, items)
		assert.Empty(t, items)
	})
}

func TestItemRepo_List(t *testing.T) {
	tests := []struct {
		name   string
		filter catalog.ItemFilter
		verify func(t *testing.T, items []catalogmodels.Item)
	}{
		{
			name:   "default filter keeps only active items ordered by sku",
			filter: catalog.ItemFilter{},
			verify: func(t *testing.T, items []catalogmodels.Item) {
				require.Len(t, items, 2)
				assert.Equal(t, "MUG-01", items[0].SKU)
				assert.Equal(t, "TEE-01", items[1].SKU)
			},
		},
		{
			name:   "include_inactive adds inactive items",
			filter: catalog.ItemFilter{IncludeInactive: true},
			verify: func(t *testing.T, items []catalogmodels.Item) {
				require.Len(t, items, 3)
			},
		},
		{
			name:   "sku filter matches a substring case-insensitively",
			filter: catalog.ItemFilter{SKU: "mug"},
			verify: func(t *testing.T, items []catalogmodels.Item) {
				require.Len(t, items, 1)
				assert.Equal(t, "MUG-01", items[0].SKU)
			},
		},
		{
			name:   "unknown sku matches nothing but stays non-nil",
			filter: catalog.ItemFilter{SKU: "NOPE"},
			verify: func(t *testing.T, items []catalogmodels.Item) {
				require.NotNil(t, items)
				assert.Empty(t, items)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := testdb.Open(t)
			repo := itemrepo.New(db)

			seedItem(t, db, teeID, "TEE-01", true)
			seedItem(t, db, mugID, "MUG-01", true)
			seedItem(t, db, deadID, "RETRO-POSTER", false)

			items, err := repo.List(context.Background(), tt.filter)

			require.NoError(t, err)
			tt.verify(t, items)
		})
	}
}
