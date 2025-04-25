package effectivemobile

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	storage "github.com/xoticdsign/effectivemobile/internal/storage/postgresql"
	"github.com/xoticdsign/effectivemobile/internal/utils/config"
)

const source = "effectivemobile"

type Service struct {
	S S

	log    *slog.Logger
	config config.EffectiveMobileConfig
}

type Handlerer interface {
	DeleteByID(id string) error
	UpdateByID(id string) error
	Create(name string, surname string, patronymic string) error
}

type S struct {
	Handlers Handlerer
}

func New(config config.EffectiveMobileConfig, storage *storage.Storage, log *slog.Logger) *Service {
	return &Service{
		S: S{
			Handlers: handlers{
				Storage: storage.DB.Handlers,

				log:    log,
				config: config,
			},
		},

		log:    log,
		config: config,
	}
}

type Querier interface {
	DeleteByID(id string) error
	UpdateByID(id string) error
	Create(name string, surname string, patronymic string) error
}

type handlers struct {
	UnimplementedHandlers

	Storage Querier

	log    *slog.Logger
	config config.EffectiveMobileConfig
}

func (h handlers) DeleteByID(id string) error {
	const op = "service.DeleteByID()"

	h.log.Debug(
		"данные получены сервисным слоем",
		slog.String("source", source),
		slog.String("op", op),
		slog.Any("data", []string{id}),
	)

	err := h.Storage.DeleteByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%s.%w", op, sql.ErrNoRows)
		}
		return fmt.Errorf("%s.%v", op, err)
	}
	h.log.Debug(
		"данные обработаны сервисным слоем",
		slog.String("source", source),
		slog.String("op", op),
	)

	return nil
}

func (h handlers) UpdateByID(id string) error {
	err := h.Storage.UpdateByID(id)
	if err != nil {
		// ERROR HANDLING
	}
	return nil
}

func (h handlers) Create(name string, surname string, patronymic string) error {
	// MAKE REQUEST TO OPEN API

	// SEND DATA TO DB

	return nil
}

// МОКИ

type UnimplementedHandlers struct{}

func (u UnimplementedHandlers) DeleteByID(id string) error {
	return nil
}

func (u UnimplementedHandlers) UpdateByID(id string) error {
	return nil
}

func (u UnimplementedHandlers) Create(name string, surname string, patronymic string) error {
	return nil
}
