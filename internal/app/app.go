package app

import (
	"github.com/cryskram/cogniq/internal/config"
	"github.com/rs/zerolog"
)

type App struct {
	Config *config.Config
	Logger zerolog.Logger
}