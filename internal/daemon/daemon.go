package daemon

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/cryskram/cogniq/internal/app"
	"github.com/cryskram/cogniq/internal/db"
)

type Daemon struct {
	app *app.App
}

func New(app *app.App) *Daemon {
	return &Daemon{
		app: app,
	}
}

func (d *Daemon) Run(ctx context.Context) error {
	if err := d.initDataDir(); err != nil {
		return err
	}

	if err := d.openDB(); err != nil {
		return err
	}
	defer d.closeDB()

	if err := db.Migrate(d.app.DB); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	d.app.Logger.Info().Msg("daemon ready")

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	<-ctx.Done()

	d.app.Logger.Info().Msg("shutting down")

	if errors.Is(ctx.Err(), context.Canceled) {
		return nil
	}
	return ctx.Err()
}

func (d *Daemon) initDataDir() error {
	dir := d.app.Config.Core.DataDir
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create data dir %s: %w", dir, err)
	}
	return nil
}

func (d *Daemon) openDB() error {
	path := filepath.Join(d.app.Config.Core.DataDir, "cogniq.db")
	database, err := db.Open(path)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	d.app.DB = database
	return nil
}

func (d *Daemon) closeDB() {
	if d.app.DB != nil {
		if err := d.app.DB.Close(); err != nil {
			d.app.Logger.Error().Err(err).Msg("close database")
		}
	}
}
