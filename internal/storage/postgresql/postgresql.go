package postgresql

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	_ "github.com/lib/pq"

	"github.com/xoticdsign/effectivemobile/internal/utils/config"
)

var (
	ErrNoNewValues              = fmt.Errorf("в запросе не были предотавлены новые данные")
	ErrOperationDidNotSuccessed = fmt.Errorf("операция не была выполнена")
)

const source = "postgresql"

type Storage struct {
	DB *DB

	log    *slog.Logger
	config config.PostgreSQLConfig
}

type Handlerer interface {
	DeleteByID(id string) error
	UpdateByID(id string, data []byte) error
	Create(name string, surname string, patronymic string, age int, gender string, nationality string) error
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
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
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
		return err
	}
	return nil
}

type Row struct {
	Name        string `json:"name"`
	Surname     string `json:"surname"`
	Patronymic  string `json:"patronymic"`
	Age         string `json:"age"`
	Gender      string `json:"gender"`
	Nationality string `json:"nationality"`
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
		return err
	}
	defer tx.Rollback()

	result, err := tx.Exec(fmt.Sprintf("DELETE FROM %s WHERE id=%s;", h.config.Table, id))
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	h.log.Debug(
		"транзакция завершена",
		slog.String("source", source),
		slog.String("op", op),
	)

	return tx.Commit()
}

func buildUpdateByIDQuery(id string, original []byte, update []byte, config config.PostgreSQLConfig) (string, error) {
	const op = "utils.buildUpdateByIDQuery()"
	var o Row
	var u Row

	json.Unmarshal(original, &o)
	json.Unmarshal(update, &u)

	count := 0
	t := []string{}

	if o.Name != u.Name && u.Name != "" {
		t = append(t, "name="+"'"+u.Name+"'")
		count++
	}

	if o.Surname != u.Surname && u.Surname != "" {
		t = append(t, "surname="+"'"+u.Surname+"'")
		count++
	}

	if o.Patronymic != u.Patronymic && u.Patronymic != "" {
		t = append(t, "patronymic="+"'"+u.Patronymic+"'")
		count++
	}

	if o.Age != u.Age && u.Age != "" {
		t = append(t, "age="+"'"+u.Age+"'")
		count++
	}

	if o.Gender != u.Gender && u.Gender != "" {
		t = append(t, "gender="+"'"+u.Gender+"'")
		count++
	}

	if o.Nationality != u.Nationality && u.Nationality != "" {
		t = append(t, "nationality="+"'"+u.Nationality+"'")
		count++
	}

	if count == 0 {
		return "", ErrNoNewValues
	}

	values := strings.Join(t, ", ")

	return fmt.Sprintf("UPDATE %s SET %s WHERE id=%s;", config.Table, values, id), nil
}

func (h handlers) UpdateByID(id string, data []byte) error {
	const op = "postgresql.UpdateByID()"

	h.log.Debug(
		"старт транзакции",
		slog.String("source", source),
		slog.String("op", op),
	)

	tx, err := h.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	r := tx.QueryRow(fmt.Sprintf("SELECT name, surname, patronymic, age, gender, nationality FROM %s WHERE id=%s;", h.config.Table, id))
	if r.Err() != nil {
		return r.Err()
	}

	var original Row

	err = r.Scan(&original.Name, &original.Surname, &original.Patronymic, &original.Age, &original.Gender, &original.Nationality)
	if err != nil {
		return err
	}

	originalByte, err := json.Marshal(original)
	if err != nil {
		return err
	}

	query, err := buildUpdateByIDQuery(id, originalByte, data, h.config)
	if err != nil {
		return ErrNoNewValues
	}

	result, err := tx.Exec(query)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrOperationDidNotSuccessed
	}

	h.log.Debug(
		"транзакция завершена",
		slog.String("source", source),
		slog.String("op", op),
	)

	return tx.Commit()
}

func (h handlers) Create(name string, surname string, patronymic string, age int, gender string, nationality string) error {
	const op = "postgresql.Create()"

	h.log.Debug(
		"старт транзакции",
		slog.String("source", source),
		slog.String("op", op),
	)

	tx, err := h.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.Exec(fmt.Sprintf("INSERT INTO %s (name, surname, patronymic, age, gender, nationality) VALUES('%s', '%s', '%s', %v, '%s', '%s');", h.config.Table, name, surname, patronymic, age, gender, nationality))
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrOperationDidNotSuccessed
	}

	h.log.Debug(
		"транзакция завершена",
		slog.String("source", source),
		slog.String("op", op),
	)

	return tx.Commit()
}

// МОКИ

type UnimplementedHandlers struct{}

func (u UnimplementedHandlers) DeleteByID(id string) error {
	return nil
}

func (u UnimplementedHandlers) UpdateByID(id string, data []byte) error {
	return nil
}

func (u UnimplementedHandlers) Create(name string, surname string, patronymic string, age int, gender string, nationality string) error {
	return nil
}
