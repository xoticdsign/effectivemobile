package effectivemobile

import (
	"log/slog"

	storage "github.com/xoticdsign/effectivemobile/internal/storage/postgresql"
	"github.com/xoticdsign/effectivemobile/internal/utils/config"
)

const source = "effectivemobile"

type Querier interface{}

type Service struct {
	UnimplementedService

	Storage Querier

	log    *slog.Logger
	config config.EffectiveMobileConfig
}

func New(config config.EffectiveMobileConfig, storage *storage.Storage, log *slog.Logger) *Service {
	return &Service{
		Storage: storage,

		log:    log,
		config: config,
	}
}

// МОКИ

type UnimplementedService struct{}
