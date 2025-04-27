package postgresql

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	_ "github.com/lib/pq"

	"github.com/xoticdsign/effectivemobile/internal/utils"
	"github.com/xoticdsign/effectivemobile/internal/utils/config"
)

var (
	ErrNoNewValues              = fmt.Errorf("в запросе не были предотавлены новые данные")
	ErrNormalization            = fmt.Errorf("ошибка нормализации")
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
	ID          int    `json:"id" example:"1"`
	Name        string `json:"name" example:"Ivan"`
	Surname     string `json:"surname" example:"Petrov"`
	Patronymic  string `json:"patronymic" example:"Ivanovich"`
	Age         int    `json:"age" example:"21"`
	Gender      string `json:"gender" example:"male"`
	Nationality string `json:"nationality" example:"RU"`
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

	n, err := utils.NormalizeInput(map[string]map[string]string{
		"name":        {u.Name: "title"},
		"surname":     {u.Surname: "title"},
		"patronymic":  {u.Patronymic: "title"},
		"gender":      {u.Gender: "lowercase"},
		"nationality": {u.Nationality: "uppercase"},
	})
	if err != nil {
		return "", ErrNormalization
	}

	name := n["name"]
	surname := n["surname"]
	patronymic := n["patronymic"]
	gender := n["gender"]
	nationality := n["nationality"]

	count := 0
	t := []string{}

	if o.Name != name && name != "" {
		t = append(t, fmt.Sprintf("name='%s'", name))
		count++
	}

	if o.Surname != surname && surname != "" {
		t = append(t, fmt.Sprintf("surname='%s'", surname))
		count++
	}

	if o.Patronymic != patronymic && patronymic != "" {
		t = append(t, fmt.Sprintf("patronymic='%s'", patronymic))
		count++
	}

	if o.Age != u.Age && u.Age != 0 {
		t = append(t, fmt.Sprintf("age=%v", u.Age))
		count++
	}

	if o.Gender != gender && gender != "" {
		t = append(t, fmt.Sprintf("gender='%s'", gender))
		count++
	}

	if o.Nationality != nationality && nationality != "" {
		t = append(t, fmt.Sprintf("nationality='%s'", nationality))
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
		if errors.Is(err, ErrNoNewValues) {
			return ErrNoNewValues
		}
		return err
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

	n, err := utils.NormalizeInput(map[string]map[string]string{
		"name":        {name: "title"},
		"surname":     {surname: "title"},
		"patronymic":  {patronymic: "title"},
		"gender":      {gender: "lowercase"},
		"nationality": {nationality: "uppercase"},
	})
	if err != nil {
		return ErrNormalization
	}

	name = n["name"]
	surname = n["surname"]
	patronymic = n["patronymic"]
	gender = n["gender"]
	nationality = n["nationality"]

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

func buildSelectQuery(id string, limit []int, filter string, value string, config config.PostgreSQLConfig) (string, error) {
	if id != "" {
		return fmt.Sprintf("SELECT * FROM %s WHERE id=%s;", config.Table, id), nil
	} else {
		if filter == "" {
			return fmt.Sprintf("SELECT * FROM %s LIMIT %v OFFSET %v;", config.Table, limit[1], limit[0]), nil
		} else {
			switch {
			case filter == "name":
				n, err := utils.NormalizeInput(map[string]map[string]string{
					"name": {value: "title"},
				})
				if err != nil {
					return "", ErrNormalization
				}
				value = n["name"]

			case filter == "surname":
				n, err := utils.NormalizeInput(map[string]map[string]string{
					"surname": {value: "title"},
				})
				if err != nil {
					return "", ErrNormalization
				}
				value = n["surname"]

			case filter == "patronymic":
				n, err := utils.NormalizeInput(map[string]map[string]string{
					"patronymic": {value: "title"},
				})
				if err != nil {
					return "", ErrNormalization
				}
				value = n["patronymic"]

			case filter == "gender":
				n, err := utils.NormalizeInput(map[string]map[string]string{
					"gender": {value: "lowercase"},
				})
				if err != nil {
					return "", ErrNormalization
				}
				value = n["gender"]

			case filter == "nationality":
				n, err := utils.NormalizeInput(map[string]map[string]string{
					"nationality": {value: "uppercase"},
				})
				if err != nil {
					return "", ErrNormalization
				}
				value = n["nationality"]
			}

			where := fmt.Sprintf("%s='%s'", filter, value)

			return fmt.Sprintf("SELECT * FROM %s WHERE %s LIMIT %v OFFSET %v;", config.Table, where, limit[1], limit[0]), nil
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

	query, err := buildSelectQuery(id, limit, filter, value, h.config)
	if err != nil {
		return nil, err
	}

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
	return []Row{}, nil
}
