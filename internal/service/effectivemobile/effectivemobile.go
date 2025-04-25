package effectivemobile

import (
	"log/slog"

	"github.com/xoticdsign/effectivemobile/internal/utils/config"
)

type Service struct {
	UnimplementedService

	log    *slog.Logger
	config config.EffectiveMobileConfig
}

func New(config config.EffectiveMobileConfig, log *slog.Logger) *Service {
	return &Service{
		log:    log,
		config: config,
	}
}

// МОКИ

type UnimplementedService struct{}
