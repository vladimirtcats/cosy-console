package order_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"cosy-console/internal/domain/orders"
	ordersmodels "cosy-console/internal/domain/orders/models"
	orderrepo "cosy-console/internal/domain/orders/repositories/order"
	"cosy-console/internal/testdb"
	"cosy-console/pkg/pagination"
)

var (
	adaUserID = int64(1)
	bobUserID = int64(2)
	mugItemID = uuid.MustParse("00000000-0000-0000-0000-0000000000a1")
	teeItemID = uuid.MustParse("00000000-0000-0000-0000-0000000000a3")
)

func seedUser(t *testing.T, db *gorm.DB, id int64, email string) {
	t.Helper()

	err := db.Exec(
		"INSERT INTO users (id, email, name) VALUES (?, ?, ?)",
		id, email, email,
	).Error
	require.NoError(t, err)
}

func seedItem(t *testing.T, db *gorm.DB, id uuid.UUID, sku string, priceCents int64) {
	t.Helper()

	err := db.Exec(
		"INSERT INTO items (id, sku, title, price_cents) VALUES (?, ?, ?, ?)",
		id, sku, sku, priceCents,
	).Error
	require.NoError(t, err)
}

func seedOrder(
	t *testing.T,
	db *gorm.DB,
	id uuid.UUID,
	userID int64,
	status string,
	createdAt time.Time,
	lines ...ordersmodels.OrderLine,
) {
	t.Helper()

	err := db.Exec(`
		INSERT INTO orders (id, user_id, status, note, total_cents, created_at)
		VALUES (?, ?, ?, '', 0, ?)`,
		id, userID, status, createdAt,
	).Error
	require.NoError(t, err)

	for _, line := range lines {
		err := db.Exec(`
			INSERT INTO order_lines (order_id, item_id, qty, price_cents)
			VALUES (?, ?, ?, ?)`,
			id, line.ItemID, line.Qty, line.PriceCents,
		).Error
		require.NoError(t, err)
	}
}

func TestOrderRepo_Get(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, db *gorm.DB) uuid.UUID
		wantErr error
		verify  func(t *testing.T, order ordersmodels.Order)
	}{
		{
			name: "returns order with preloaded lines",
			setup: func(t *testing.T, db *gorm.DB) uuid.UUID {
				seedUser(t, db, adaUserID, "ada@example.com")
				seedItem(t, db, mugItemID, "MUG-01", 990)

				id := uuid.New()
				seedOrder(t, db, id, adaUserID, ordersmodels.StatusPending, time.Now(),
					ordersmodels.OrderLine{ItemID: mugItemID, Qty: 2, PriceCents: 990},
				)

				return id
			},
			verify: func(t *testing.T, order ordersmodels.Order) {
				assert.Equal(t, adaUserID, order.UserID)
				assert.Equal(t, ordersmodels.StatusPending, order.Status)
				require.Len(t, order.Lines, 1)
				assert.Equal(t, mugItemID, order.Lines[0].ItemID)
				assert.Equal(t, 2, order.Lines[0].Qty)
			},
		},
		{
			name: "missing order maps to the domain ErrNotFound sentinel",
			setup: func(t *testing.T, _ *gorm.DB) uuid.UUID {
				return uuid.New()
			},
			wantErr: orders.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := testdb.Open(t)
			repo := orderrepo.New(db)

			id := tt.setup(t, db)

			order, err := repo.Get(context.Background(), id)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)

				return
			}

			require.NoError(t, err)
			tt.verify(t, order)
		})
	}
}

func TestOrderRepo_ListByUser(t *testing.T) {
	older := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
	newer := time.Date(2026, 9, 2, 10, 0, 0, 0, time.UTC)
	newest := time.Date(2026, 9, 3, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name   string
		setup  func(t *testing.T, db *gorm.DB)
		filter orders.OrderFilter
		opts   []pagination.Option
		userID int64
		verify func(t *testing.T, ordersList []ordersmodels.Order, result pagination.Result)
	}{
		{
			name: "returns only the user orders, newest first, with the page result",
			setup: func(t *testing.T, db *gorm.DB) {
				seedUser(t, db, adaUserID, "ada@example.com")
				seedUser(t, db, bobUserID, "bob@example.com")

				seedOrder(t, db, uuid.New(), adaUserID, ordersmodels.StatusPending, older)
				seedOrder(t, db, uuid.New(), adaUserID, ordersmodels.StatusPaid, newest)
				seedOrder(t, db, uuid.New(), adaUserID, ordersmodels.StatusPending, newer)
				seedOrder(t, db, uuid.New(), bobUserID, ordersmodels.StatusPaid, newest)
			},
			userID: adaUserID,
			verify: func(t *testing.T, ordersList []ordersmodels.Order, result pagination.Result) {
				require.Len(t, ordersList, 3)
				assert.Equal(t, ordersmodels.StatusPaid, ordersList[0].Status, "newest first")
				assert.Equal(t, int64(3), result.Total)
				assert.Equal(t, 1, result.Page)
			},
		},
		{
			name: "status filter keeps only matching orders",
			setup: func(t *testing.T, db *gorm.DB) {
				seedUser(t, db, adaUserID, "ada@example.com")

				seedOrder(t, db, uuid.New(), adaUserID, ordersmodels.StatusPending, older)
				seedOrder(t, db, uuid.New(), adaUserID, ordersmodels.StatusPaid, newer)
			},
			userID: adaUserID,
			filter: orders.OrderFilter{Status: ordersmodels.StatusPaid},
			verify: func(t *testing.T, ordersList []ordersmodels.Order, result pagination.Result) {
				require.Len(t, ordersList, 1)
				assert.Equal(t, ordersmodels.StatusPaid, ordersList[0].Status)
				assert.Equal(t, int64(1), result.Total)
			},
		},
		{
			name: "second page serves the remaining orders",
			setup: func(t *testing.T, db *gorm.DB) {
				seedUser(t, db, adaUserID, "ada@example.com")

				seedOrder(t, db, uuid.New(), adaUserID, ordersmodels.StatusPending, older)
				seedOrder(t, db, uuid.New(), adaUserID, ordersmodels.StatusPaid, newer)
				seedOrder(t, db, uuid.New(), adaUserID, ordersmodels.StatusShipped, newest)
			},
			userID: adaUserID,
			opts:   []pagination.Option{pagination.WithPage(2), pagination.WithPageSize(2)},
			verify: func(t *testing.T, ordersList []ordersmodels.Order, result pagination.Result) {
				require.Len(t, ordersList, 1)
				assert.Equal(t, ordersmodels.StatusPending, ordersList[0].Status, "oldest row is on page 2")
				assert.Equal(t, int64(3), result.Total)
				assert.Equal(t, 2, result.TotalPages)
			},
		},
		{
			name: "user without orders gets an empty slice, not nil",
			setup: func(t *testing.T, db *gorm.DB) {
				seedUser(t, db, adaUserID, "ada@example.com")
			},
			userID: bobUserID,
			verify: func(t *testing.T, ordersList []ordersmodels.Order, result pagination.Result) {
				require.NotNil(t, ordersList)
				assert.Empty(t, ordersList)
				assert.Equal(t, int64(0), result.Total)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := testdb.Open(t)
			repo := orderrepo.New(db)

			tt.setup(t, db)

			ordersList, result, err := repo.ListByUser(context.Background(), tt.userID, tt.filter, tt.opts...)

			require.NoError(t, err)
			tt.verify(t, ordersList, result)
		})
	}
}

func TestOrderRepo_Create(t *testing.T) {
	t.Run("persists the order with its lines", func(t *testing.T) {
		db := testdb.Open(t)
		repo := orderrepo.New(db)

		seedUser(t, db, adaUserID, "ada@example.com")
		seedItem(t, db, mugItemID, "MUG-01", 990)
		seedItem(t, db, teeItemID, "TEE-01", 1490)

		order := ordersmodels.Order{
			ID:     uuid.New(),
			UserID: adaUserID,
			Status: ordersmodels.StatusPending,
			Lines: []ordersmodels.OrderLine{
				{ItemID: mugItemID, Qty: 1, PriceCents: 990},
			},
		}
		order.Lines[0].OrderID = order.ID

		err := repo.Create(context.Background(), order)

		require.NoError(t, err)

		stored, err := repo.Get(context.Background(), order.ID)
		require.NoError(t, err)
		assert.Equal(t, ordersmodels.StatusPending, stored.Status)
		require.Len(t, stored.Lines, 1)
		assert.Equal(t, mugItemID, stored.Lines[0].ItemID)
	})
}

func TestOrderRepo_Update(t *testing.T) {
	note := "gift wrap"
	status := ordersmodels.StatusPaid

	tests := []struct {
		name    string
		attrs   orders.UpdateAttrs
		wantErr error
		verify  func(t *testing.T, before, after ordersmodels.Order)
	}{
		{
			name:  "updates the note and returns the fresh row",
			attrs: orders.UpdateAttrs{Note: &note},
			verify: func(t *testing.T, before, after ordersmodels.Order) {
				assert.Equal(t, "gift wrap", after.Note)
				assert.Equal(t, before.Status, after.Status, "untouched fields keep their values")
			},
		},
		{
			name:  "updates the status",
			attrs: orders.UpdateAttrs{Status: &status},
			verify: func(t *testing.T, before, after ordersmodels.Order) {
				assert.Equal(t, ordersmodels.StatusPaid, after.Status)
				assert.Equal(t, before.Note, after.Note)
			},
		},
		{
			name:  "all-nil attrs is a no-op returning the current row",
			attrs: orders.UpdateAttrs{},
			verify: func(t *testing.T, before, after ordersmodels.Order) {
				assert.Equal(t, before, after)
			},
		},
		{
			name:    "missing order maps to the domain ErrNotFound sentinel",
			attrs:   orders.UpdateAttrs{Note: &note},
			wantErr: orders.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := testdb.Open(t)
			repo := orderrepo.New(db)

			seedUser(t, db, adaUserID, "ada@example.com")
			seedItem(t, db, mugItemID, "MUG-01", 990)

			id := uuid.New()
			seedOrder(t, db, id, adaUserID, ordersmodels.StatusPending, time.Now(),
				ordersmodels.OrderLine{ItemID: mugItemID, Qty: 1, PriceCents: 990},
			)

			target := id
			if tt.wantErr != nil {
				target = uuid.New() // not seeded
			}

			before, err := repo.Get(context.Background(), id)
			if tt.wantErr == nil {
				require.NoError(t, err)
			}

			after, err := repo.Update(context.Background(), target, tt.attrs)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)

				return
			}

			require.NoError(t, err)
			tt.verify(t, before, after)
		})
	}
}
