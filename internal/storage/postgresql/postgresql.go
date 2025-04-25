package postgresql

import (
	"database/sql"
	"fmt"
	"log/slog"

	_ "github.com/lib/pq"
	"github.com/xoticdsign/effectivemobile/internal/utils/config"
)

const source = "postgresql"

type Storage struct {
	DB DB

	log    *slog.Logger
	config config.PostgreSQLConfig
}

type Handlerer interface {
	DeleteByID(id string) error
	UpdateByID(id string) error
	Create(name string, surname string, patronymic string) error
}

type DB struct {
	Implementation *sql.DB
	Handlers       Handlerer
}

func New(config config.PostgreSQLConfig, log *slog.Logger) (*Storage, error) {
	const op = "postgresql.New()"

	db, err := sql.Open("postgres", config.Address)
	if err != nil {
		return nil, fmt.Errorf("%s @ %v", op, err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("%s @ %v", op, err)
	}

	return &Storage{
		DB: DB{
			Implementation: db,
			Handlers: handlers{
				log:    log,
				config: config,
			},
		},

		log:    log,
		config: config,
	}, nil
}

func (s *Storage) Shutdown() error {
	const op = "postgresql.Shutdown()"

	err := s.DB.Implementation.Close()
	if err != nil {
		return fmt.Errorf("%s @ %v", op, err)
	}
	return nil
}

type handlers struct {
	UnimplementedHandlers

	log    *slog.Logger
	config config.PostgreSQLConfig
}

func (h handlers) DeleteByID(id string) error {
	// CALL TO DB

	return nil
}

func (h handlers) UpdateByID(id string) error {
	// CALL TO DB

	return nil
}

func (h handlers) Create(name string, surname string, patronymic string) error {
	// CALL TO DB

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
	// CALL TO DB

	return nil
}
