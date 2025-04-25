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
	DB *sql.DB

	log    *slog.Logger
	config config.PostgreSQLConfig
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
		DB: db,

		log:    log,
		config: config,
	}, nil
}

func (s *Storage) Shutdown() error {
	const op = "postgresql.Shutdown()"

	err := s.DB.Close()
	if err != nil {
		return fmt.Errorf("%s @ %v", op, err)
	}
	return nil
}
