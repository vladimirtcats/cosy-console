package db

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	syncOnce sync.Once
	db       *gorm.DB
)

type Option func(*connectOptions)

type connectOptions struct {
	readOnly bool
}

// WithReadOnly asks the server to reject every write at the protocol level
// (default_transaction_read_only=on in the startup message).
func WithReadOnly() Option {
	return func(o *connectOptions) {
		o.readOnly = true
	}
}

func Connect(dsn string, opts ...Option) *gorm.DB {
	syncOnce.Do(func() {
		db = connect(dsn, opts...)
	})

	return db
}

func connect(dsn string, opts ...Option) *gorm.DB {
	options := connectOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	pgxConfig, err := pgx.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("parse postgres config: %v", err)
	}

	pgxConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	pgxConfig.RuntimeParams["TimeZone"] = "UTC"
	if options.readOnly {
		pgxConfig.RuntimeParams["default_transaction_read_only"] = "on"
	}

	conn, err := gorm.Open(
		postgres.New(postgres.Config{
			DSN:        stdlib.RegisterConnConfig(pgxConfig),
			DriverName: "pgx",
		}),
		&gorm.Config{TranslateError: true},
	)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}

	sqlDB, err := conn.DB()
	if err != nil {
		log.Fatalf("get sql.DB: %v", err)
	}

	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(20 * time.Minute)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("ping database: %v", err)
	}

	// The timezone and read-only pins are startup parameters, and a pooler
	// may silently strip them. Ask the server what it actually got.
	var timeZone string
	if err := conn.Raw("SHOW timezone").Scan(&timeZone).Error; err != nil {
		log.Fatalf("read timezone: %v", err)
	}
	if err := validateTimeZone(timeZone); err != nil {
		log.Fatalf("database connection check failed: %v", err)
	}

	if options.readOnly {
		var readOnly string
		if err := conn.Raw("SHOW default_transaction_read_only").Scan(&readOnly).Error; err != nil {
			log.Fatalf("read default_transaction_read_only: %v", err)
		}
		if err := validateReadOnly(readOnly); err != nil {
			log.Fatalf("database connection check failed: %v", err)
		}
	}

	log.Printf("database connection is ok, timezone: %s", strings.TrimSpace(timeZone))

	return conn
}

func validateTimeZone(value string) error {
	if !strings.EqualFold(strings.TrimSpace(value), "UTC") {
		return fmt.Errorf("connection timezone is %q, want UTC: "+
			"the TimeZone=UTC startup parameter was overridden or stripped "+
			"(check DATABASE_URL and pooler ignore_startup_parameters)", value)
	}

	return nil
}

func validateReadOnly(value string) error {
	if !strings.EqualFold(strings.TrimSpace(value), "on") {
		return fmt.Errorf("default_transaction_read_only is %q, want on: "+
			"the read-only startup parameter was overridden or stripped "+
			"(check DATABASE_URL and pooler ignore_startup_parameters)", value)
	}

	return nil
}
