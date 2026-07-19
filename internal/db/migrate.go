package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/cryskram/cogniq/sql/migrations"
	"github.com/pressly/goose/v3"
)

func Migrate(db *sql.DB) error {
	provider, err := goose.NewProvider(
		goose.DialectSQLite3,
		db,
		migrations.Content,
		goose.WithDisableGlobalRegistry(true),
	)
	if err != nil {
		return fmt.Errorf("goose provider: %w", err)
	}

	if _, err := provider.Up(context.Background()); err != nil {
		return fmt.Errorf("migration up: %w", err)
	}

	return nil
}