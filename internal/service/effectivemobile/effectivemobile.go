package effectivemobile

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/xoticdsign/effectivemobile/internal/client"
	storage "github.com/xoticdsign/effectivemobile/internal/storage/postgresql"
	"github.com/xoticdsign/effectivemobile/internal/utils/config"
)

var (
	ErrClientNotFound  = fmt.Errorf("у клиента нет данных")
	ErrClientInternal  = fmt.Errorf("внутренняя ошибка клиента")
	ErrStorageNotFound = fmt.Errorf("у хранилища нет данных")
	ErrStorageConflict = fmt.Errorf("запрос сформирофан некоректно")
	ErrStorageInternal = fmt.Errorf("внутренняя ошибка хранилища")
)

const source = "service"

type Service struct {
	S S

	log    *slog.Logger
	config config.EffectiveMobileConfig
}

type Handlerer interface {
	DeleteByID(id string) error
	UpdateByID(id string, data []byte) error
	Create(name string, surname string, patronymic string) error
	Select(id string, limit []int, filter string, value string) ([]storage.Row, error)
}

type S struct {
	Handlers Handlerer
}

func New(config config.EffectiveMobileConfig, client *client.Client, storage *storage.Storage, log *slog.Logger) *Service {
	return &Service{
		S: S{
			Handlers: Handlers{
				Client:  client.C.Handlers,
				Storage: storage.DB.Handlers,

				log:    log,
				config: config,
			},
		},

		log:    log,
		config: config,
	}
}

type Clienter interface {
	GetAge(name string) (int, error)
	GetGender(name string) (string, error)
	GetNationality(name string) (string, error)
}

type Querier interface {
	DeleteByID(id string) error
	UpdateByID(id string, data []byte) error
	Create(name string, surname string, patronymic string, age int, gender string, nationality string) error
	Select(id string, limit []int, filter string, value string) ([]storage.Row, error)
}

type Handlers struct {
	UnimplementedHandlers

	Client  Clienter
	Storage Querier

	log    *slog.Logger
	config config.EffectiveMobileConfig
}

func (h Handlers) DeleteByID(id string) error {
	const op = "service.DeleteByID()"

	h.log.Debug(
		"данные получены сервисным слоем",
		slog.String("source", source),
		slog.String("op", op),
	)

	err := h.Storage.DeleteByID(id)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			h.log.Error(
				"в хранилище нет соответсвующих данных",
				slog.String("source", source),
				slog.String("op", op),
				slog.Any("error", err),
			)

			return fmt.Errorf("%w: %v", ErrStorageNotFound, err)

		default:
			h.log.Error(
				"внутренняя ошибка хранилища",
				slog.String("source", source),
				slog.String("op", op),
				slog.Any("error", err),
			)

			return fmt.Errorf("%w: %v", ErrStorageInternal, err)
		}
	}
	h.log.Debug(
		"данные обработаны сервисным слоем",
		slog.String("source", source),
		slog.String("op", op),
	)

	return nil
}

func (h Handlers) UpdateByID(id string, data []byte) error {
	const op = "service.UpdateByID()"

	h.log.Debug(
		"данные получены сервисным слоем",
		slog.String("source", source),
		slog.String("op", op),
	)

	err := h.Storage.UpdateByID(id, data)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrConstraint):
			h.log.Error(
				"предоставленные данные не соответсвуют ограничениям таблицы",
				slog.String("source", source),
				slog.String("op", op),
				slog.Any("error", err),
			)

			return fmt.Errorf("%w: %v", ErrStorageConflict, err)

		case errors.Is(err, sql.ErrNoRows):
			h.log.Error(
				"в хранилище нет соответсвующих данных",
				slog.String("source", source),
				slog.String("op", op),
				slog.Any("error", err),
			)

			return fmt.Errorf("%w: %v", ErrStorageNotFound, err)

		case errors.Is(err, storage.ErrNoNewValues):
			h.log.Error(
				"хранилище не обнаружило новых данных в запросе",
				slog.String("source", source),
				slog.String("op", op),
				slog.Any("error", err),
			)

			return fmt.Errorf("%w: %v", ErrStorageConflict, err)

		default:
			h.log.Error(
				"внутренняя ошибка хранилища",
				slog.String("source", source),
				slog.String("op", op),
				slog.Any("error", err),
			)

			return fmt.Errorf("%w: %v", ErrStorageInternal, err)
		}
	}
	h.log.Debug(
		"данные обработаны сервисным слоем",
		slog.String("source", source),
		slog.String("op", op),
	)

	return nil
}

func (h Handlers) Create(name string, surname string, patronymic string) error {
	const op = "service.Create()"

	h.log.Debug(
		"данные получены сервисным слоем",
		slog.String("source", source),
		slog.String("op", op),
	)

	errChan := make(chan error, 3)

	var age int
	var gender string
	var nationality string

	wg := sync.WaitGroup{}

	wg.Add(1)

	go func() {
		var err error

		age, err = h.Client.GetAge(name)
		if err != nil {
			errChan <- err
		}

		wg.Done()
	}()

	wg.Add(1)

	go func() {
		var err error

		gender, err = h.Client.GetGender(name)
		if err != nil {
			errChan <- err
		}

		wg.Done()
	}()

	wg.Add(1)

	go func() {
		var err error

		nationality, err = h.Client.GetNationality(name)
		if err != nil {
			errChan <- err
		}

		wg.Done()
	}()

	wg.Wait()

	close(errChan)

	for e := range errChan {
		if errors.Is(e, client.ErrNotFound) {
			h.log.Error(
				"клиент не ничего нашел",
				slog.String("source", source),
				slog.String("op", op),
				slog.Any("error", e),
			)

			return fmt.Errorf("%w: %v", ErrClientNotFound, e)
		} else {
			h.log.Error(
				"внутренняя ошибка клиента",
				slog.String("source", source),
				slog.String("op", op),
				slog.Any("error", e),
			)

			return fmt.Errorf("%w: %v", ErrClientInternal, e)
		}
	}

	err := h.Storage.Create(name, surname, patronymic, age, gender, nationality)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			h.log.Error(
				"в хранилище нет соответсвующих данных",
				slog.String("source", source),
				slog.String("op", op),
				slog.Any("error", err),
			)

			return fmt.Errorf("%w: %v", ErrStorageNotFound, err)

		default:
			h.log.Error(
				"внутренняя ошибка хранилища",
				slog.String("source", source),
				slog.String("op", op),
				slog.Any("error", err),
			)

			return fmt.Errorf("%w: %v", ErrStorageInternal, err)
		}
	}
	h.log.Debug(
		"данные обработаны сервисным слоем",
		slog.String("source", source),
		slog.String("op", op),
	)

	return nil
}

func (h Handlers) Select(id string, limit []int, filter string, value string) ([]storage.Row, error) {
	const op = "service.Select()"

	h.log.Debug(
		"данные получены сервисным слоем",
		slog.String("source", source),
		slog.String("op", op),
	)

	r, err := h.Storage.Select(id, limit, filter, value)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			h.log.Error(
				"в хранилище нет соответсвующих данных",
				slog.String("source", source),
				slog.String("op", op),
				slog.Any("error", err),
			)

			return nil, fmt.Errorf("%w: %v", ErrStorageNotFound, err)

		default:
			h.log.Error(
				"внутренняя ошибка хранилища",
				slog.String("source", source),
				slog.String("op", op),
				slog.Any("error", err),
			)

			return nil, fmt.Errorf("%w: %v", ErrStorageInternal, err)
		}
	}
	h.log.Debug(
		"данные обработаны сервисным слоем",
		slog.String("source", source),
		slog.String("op", op),
	)

	return r, nil
}

// МОКИ

type UnimplementedHandlers struct{}

func (u UnimplementedHandlers) DeleteByID(id string) error {
	switch id {
	case "s404":
		return ErrStorageNotFound

	case "s500":
		return ErrStorageInternal
	}
	return nil
}

func (u UnimplementedHandlers) UpdateByID(id string, data []byte) error {
	switch id {
	case "s404":
		return ErrStorageNotFound

	case "s409":
		return ErrStorageConflict

	case "s500":
		return ErrStorageInternal
	}
	return nil
}

func (u UnimplementedHandlers) Create(name string, surname string, patronymic string) error {
	switch name {
	case "s404":
		return ErrStorageNotFound

	case "c404":
		return ErrClientNotFound

	case "500":
		return ErrStorageInternal
	}
	return nil
}

func (u UnimplementedHandlers) Select(id string, limit []int, filter string, value string) ([]storage.Row, error) {
	switch id {
	case "s404":
		return []storage.Row{}, ErrStorageNotFound

	case "s500":
		return []storage.Row{}, ErrStorageInternal
	}
	return []storage.Row{}, nil
}
