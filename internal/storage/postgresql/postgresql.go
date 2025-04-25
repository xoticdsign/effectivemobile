package postgresql

import (
	"database/sql"
	"fmt"
	"log/slog"

	_ "github.com/lib/pq"

	"github.com/xoticdsign/effectivemobile/internal/utils"
	"github.com/xoticdsign/effectivemobile/internal/utils/config"
)

const source = "postgresql"

type Storage struct {
	DB *DB

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

	var conn string

	switch {
	case config.Password == "":
		conn = fmt.Sprintf("postgres://%s@%s:%s/%s?sslmode=%s&&%s", config.Username, config.Host, config.Port, config.Database, config.SSL, config.Extra)

	case config.Extra == "":
		conn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", config.Username, config.Password, config.Host, config.Port, config.Database, config.SSL)

	case config.Password == "" && config.Extra == "":
		conn = fmt.Sprintf("postgres://%s@%s:%s/%s?sslmode=%s", config.Username, config.Host, config.Port, config.Database, config.SSL)

	default:
		conn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s&&%s", config.Username, config.Password, config.Host, config.Port, config.Database, config.SSL, config.Extra)
	}

	db, err := sql.Open("postgres", conn)
	if err != nil {
		return nil, fmt.Errorf("%s @ %v", op, err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("%s @ %v", op, err)
	}

	return &Storage{
		DB: &DB{
			Implementation: db,
			Handlers: handlers{
				DB: db,

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

	DB *sql.DB

	log    *slog.Logger
	config config.PostgreSQLConfig
}

func (h handlers) DeleteByID(id string) error {
	const op = "postgresql.DeleteByID()"

	h.log.Debug(
		"старт транзакции",
		slog.String("source", source),
		slog.String("op", op),
		slog.Any("data", []string{id}),
	)

	tx, err := h.DB.Begin()
	if err != nil {
		utils.LogDBDebug(h.log, source, op, err)

		return fmt.Errorf("%s @ %w", op, err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("DELETE FROM ? WHERE id=?;")
	if err != nil {
		utils.LogDBDebug(h.log, source, op, err)

		return fmt.Errorf("%s @ %w", op, err)
	}
	defer tx.Stmt(stmt).Close()

	result, err := tx.Stmt(stmt).Exec(h.config.Database, id)
	if err != nil {
		utils.LogDBDebug(h.log, source, op, err)

		return fmt.Errorf("%s @ %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		utils.LogDBDebug(h.log, source, op, err)

		return fmt.Errorf("%s @ %w", op, err)
	}

	if rowsAffected == 0 {
		utils.LogDBDebug(h.log, source, op, err)

		return sql.ErrNoRows
	}

	h.log.Debug(
		"транзакция завершена",
		slog.String("source", source),
		slog.String("op", op),
		slog.Any("result", result),
	)

	return tx.Commit()
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
	return nil
}
