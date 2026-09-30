// Package testdb gives repository tests a database connection. Each test
// case runs inside its own transaction that rolls back on cleanup, so
// parallel test packages can share one database without seeing each
// other's rows.
package testdb

import (
	"os"
	"testing"

	"gorm.io/gorm"

	dbinfra "cosy-console/internal/infrastructure/db"
)

func Open(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL is not set: skipping the database test")
	}

	db := dbinfra.Connect(dsn)

	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("begin test transaction: %v", tx.Error)
	}

	t.Cleanup(func() {
		if err := tx.Rollback().Error; err != nil {
			t.Errorf("rollback test transaction: %v", err)
		}
	})

	// A known starting state, local to this transaction: the truncate is
	// rolled back together with everything the test seeds and writes.
	err := tx.Exec(`
		TRUNCATE order_lines, orders, items, users
		RESTART IDENTITY CASCADE
	`).Error
	if err != nil {
		t.Fatalf("truncate tables: %v", err)
	}

	return tx
}
