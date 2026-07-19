package main

import (
	"context"
	"log"

	"github.com/cryskram/cogniq/internal/app"
	"github.com/cryskram/cogniq/internal/config"
	"github.com/cryskram/cogniq/internal/daemon"
	"github.com/cryskram/cogniq/internal/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	zl := logger.New(cfg.Log)
	application := &app.App{Config: cfg, Logger: zl}

	d := daemon.New(application)
	if err := d.Run(context.Background()); err != nil {
		zl.Fatal().Err(err).Msg("daemon exited with error")
	}
}
