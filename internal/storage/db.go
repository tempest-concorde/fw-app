package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/extra/bundebug"
	"github.com/uptrace/bun/migrate"

	_ "modernc.org/sqlite"

	"github.com/tempest-concorde/fw-app/internal/storage/migrations"
)

// DB wraps bun.DB with lifecycle management
type DB struct {
	*bun.DB
	migrator *migrate.Migrator
}

// Config holds database configuration
type Config struct {
	Path        string
	Development bool
}

// New creates a new database connection
func New(cfg Config) (*DB, error) {
	// Open SQLite database
	sqldb, err := sql.Open("sqlite", cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool for SQLite
	sqldb.SetMaxOpenConns(1)

	// Create Bun DB
	db := bun.NewDB(sqldb, sqlitedialect.New())

	// Add query logging in development mode
	if cfg.Development {
		db.AddQueryHook(bundebug.NewQueryHook(
			bundebug.WithVerbose(true),
			bundebug.FromEnv("BUNDEBUG"),
		))
	}

	// Create migrator
	migrator := migrate.NewMigrator(db, migrations.Migrations)

	return &DB{
		DB:       db,
		migrator: migrator,
	}, nil
}

// Init initializes the database by running migrations
func (db *DB) Init(ctx context.Context) error {
	// Create migration tables
	if err := db.migrator.Init(ctx); err != nil {
		return fmt.Errorf("failed to init migrator: %w", err)
	}

	// Run migrations
	if err := db.migrator.Lock(ctx); err != nil {
		return fmt.Errorf("failed to lock migrator: %w", err)
	}
	defer func() {
		if unlockErr := db.migrator.Unlock(ctx); unlockErr != nil {
			fmt.Fprintf(os.Stderr, "failed to unlock migrator: %v\n", unlockErr)
		}
	}()

	group, err := db.migrator.Migrate(ctx)
	if err != nil {
		return fmt.Errorf("failed to migrate: %w", err)
	}

	if group.IsZero() {
		fmt.Fprintln(os.Stdout, "no new migrations to run")
	} else {
		fmt.Fprintf(os.Stdout, "migrated to %s\n", group)
	}

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}
