package postgresql

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"unicode"

	_ "github.com/lib/pq"

	"github.com/xoticdsign/effectivemobile/internal/utils/config"
)

var (
	ErrNoNewValues              = fmt.Errorf("в запросе не были предотавлены новые данные")
	ErrConstraint               = fmt.Errorf("был нарушен check")
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
	Select(id string, limit []int, filter string, value string) ([]Row, error)
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
			Handlers: Handlers{
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
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Surname     string `json:"surname"`
	Patronymic  string `json:"patronymic"`
	Age         int    `json:"age"`
	Gender      string `json:"gender"`
	Nationality string `json:"nationality"`
}

type Handlers struct {
	UnimplementedHandlers

	DB *sql.DB

	log    *slog.Logger
	config config.PostgreSQLConfig
}

func (h Handlers) DeleteByID(id string) error {
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
		t = append(t, fmt.Sprintf("name='%s'", u.Name))
		count++
	}

	if o.Surname != u.Surname && u.Surname != "" {
		t = append(t, fmt.Sprintf("surname='%s'", u.Surname))
		count++
	}

	if o.Patronymic != u.Patronymic && u.Patronymic != "" {
		t = append(t, fmt.Sprintf("patronymic='%s'", u.Patronymic))
		count++
	}

	if o.Age != u.Age && u.Age != 0 {
		t = append(t, fmt.Sprintf("age=%v", u.Age))
		count++
	}

	if o.Gender != u.Gender && u.Gender != "" {
		t = append(t, fmt.Sprintf("gender='%s'", u.Gender))
		count++
	}

	if o.Nationality != u.Nationality && u.Nationality != "" {
		t = append(t, fmt.Sprintf("nationality='%s'", u.Nationality))
		count++
	}

	if count == 0 {
		return "", ErrNoNewValues
	}

	values := strings.Join(t, ", ")

	return fmt.Sprintf("UPDATE %s SET %s WHERE id=%s;", config.Table, values, id), nil
}

func (h Handlers) UpdateByID(id string, data []byte) error {
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
		return ErrConstraint
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

func (h Handlers) Create(name string, surname string, patronymic string, age int, gender string, nationality string) error {
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

func buildSelectQuery(id string, limit []int, filter string, value string, config config.PostgreSQLConfig) string {
	if id != "" {
		return fmt.Sprintf("SELECT * FROM %s WHERE id=%s;", config.Table, id)
	} else {
		if filter == "" {
			return fmt.Sprintf("SELECT * FROM %s LIMIT %v OFFSET %v;", config.Table, limit[1], limit[0])
		} else {
			switch {
			case filter == "name" || filter == "surname" || filter == "patronymic":
				runes := []rune(value)
				runes[0] = unicode.ToUpper(runes[0])

				for i := 1; i < len(runes); i++ {
					runes[i] = unicode.ToLower(runes[i])
				}

				value = string(runes)

			case filter == "gender":
				value = strings.ToLower(value)

			case filter == "nationality":
				value = strings.ToUpper(value)
			}

			where := fmt.Sprintf("%s='%s'", filter, value)

			return fmt.Sprintf("SELECT * FROM %s WHERE %s LIMIT %v OFFSET %v;", config.Table, where, limit[1], limit[0])
		}
	}
}

func (h Handlers) Select(id string, limit []int, filter string, value string) ([]Row, error) {
	const op = "postgresql.Select()"

	h.log.Debug(
		"старт транзакции",
		slog.String("source", source),
		slog.String("op", op),
	)

	tx, err := h.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := buildSelectQuery(id, limit, filter, value, h.config)

	r, err := tx.Query(query)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var rows []Row

	for r.Next() {
		var row Row

		err := r.Scan(&row.ID, &row.Name, &row.Surname, &row.Patronymic, &row.Age, &row.Gender, &row.Nationality)
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}

	if r.Err() != nil {
		return nil, r.Err()
	}

	if len(rows) == 0 {
		return nil, sql.ErrNoRows
	}

	h.log.Debug(
		"транзакция завершена",
		slog.String("source", source),
		slog.String("op", op),
	)

	return rows, nil
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

func (u UnimplementedHandlers) Select(id string, limit []int, filter string, value string) ([]Row, error) {
	return nil, nil
}
