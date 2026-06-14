package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		// Create samples table
		_, err := db.ExecContext(ctx, `
			CREATE TABLE IF NOT EXISTS samples (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL,
				description TEXT,
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			)
		`)
		if err != nil {
			return fmt.Errorf("failed to create samples table: %w", err)
		}

		// Create settings table
		_, err = db.ExecContext(ctx, `
			CREATE TABLE IF NOT EXISTS settings (
				key TEXT PRIMARY KEY,
				value TEXT NOT NULL,
				updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			)
		`)
		if err != nil {
			return fmt.Errorf("failed to create settings table: %w", err)
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		// Rollback: drop tables
		_, err := db.ExecContext(ctx, `DROP TABLE IF EXISTS settings`)
		if err != nil {
			return fmt.Errorf("failed to drop settings table: %w", err)
		}

		_, err = db.ExecContext(ctx, `DROP TABLE IF EXISTS samples`)
		if err != nil {
			return fmt.Errorf("failed to drop samples table: %w", err)
		}

		return nil
	})
}
