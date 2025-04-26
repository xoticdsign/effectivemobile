package effectivemobile

import (
	"errors"
	"log/slog"
	"net"

	"github.com/gofiber/fiber/v2"

	"github.com/xoticdsign/effectivemobile/internal/client"
	effectivemobileservice "github.com/xoticdsign/effectivemobile/internal/service/effectivemobile"
	storage "github.com/xoticdsign/effectivemobile/internal/storage/postgresql"
	"github.com/xoticdsign/effectivemobile/internal/utils/config"
)

const source = "effectivemobile"

type App struct {
	Server Server
	Client *client.Client

	log    *slog.Logger
	config config.EffectiveMobileConfig
}

type Handlerer interface {
	DeleteByID(c *fiber.Ctx) error
	UpdateByID(c *fiber.Ctx) error
	Create(c *fiber.Ctx) error
	Select(c *fiber.Ctx) error
}

type Server struct {
	Implementation *fiber.App
	Handlers       Handlerer
}

func New(config config.EffectiveMobileConfig, storage *storage.Storage, log *slog.Logger) *App {
	f := fiber.New(fiber.Config{
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			var e *fiber.Error

			if errors.As(err, &e) {
				return c.JSON(e)
			}
			return c.JSON(fiber.ErrInternalServerError)
		},
		AppName: "effectivemobile",
	})

	client := client.New(config, log)

	emservice := effectivemobileservice.New(config, client, storage, log)

	h := handlers{
		Service: emservice.S.Handlers,

		log:    log,
		config: config,
	}

	f.Delete("/delete/:id", h.DeleteByID)
	f.Put("/update/:id", h.UpdateByID)
	f.Post("/create", h.Create)
	f.Get("/select/:id?", h.Select)

	return &App{
		Server: Server{
			Implementation: f,
			Handlers:       h,
		},
		Client: client,

		log:    log,
		config: config,
	}
}

func (a *App) Run() error {
	const op = "effectivemobile.Run()"

	err := a.Server.Implementation.Listen(net.JoinHostPort(a.config.Host, a.config.Port))
	if err != nil {
		return err
	}
	return nil
}

func (a *App) Shutdown() error {
	const op = "effectivemobile.Shutdown()"

	err := a.Server.Implementation.Shutdown()
	if err != nil {
		return err
	}

	a.Client.Shutdown()

	return nil
}

type Servicer interface {
	DeleteByID(id string) error
	UpdateByID(id string, data []byte) error
	Create(name string, surname string, patronymic string) error
	Select(id string, limit []int, filter string, value string) ([]storage.Row, error)
}

type handlers struct {
	UnimplementedHandlers

	Service Servicer

	log    *slog.Logger
	config config.EffectiveMobileConfig
}

type DeleteByIDRequest struct{}

type DeleteByIDResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func (h handlers) DeleteByID(c *fiber.Ctx) error {
	const op = "effectivemobile.DeleteByID()"

	id := c.Params("id")
	if id == "" {
		h.log.Debug(
			"отсутсвуют параметры",
			slog.String("source", source),
			slog.String("op", op),
			slog.String("error", "absent parameters"),
		)

		return fiber.ErrBadRequest
	}

	h.log.Debug(
		"получен запрос на удаление",
		slog.String("source", source),
		slog.String("op", op),
		slog.Any("parameters", []string{id}),
	)

	err := h.Service.DeleteByID(id)
	if err != nil {
		switch {
		case errors.Is(err, effectivemobileservice.ErrStorageNotFound):
			return fiber.ErrNotFound

		default:
			return fiber.ErrInternalServerError
		}
	}
	h.log.Debug(
		"обработан запрос на удаление",
		slog.String("source", source),
		slog.String("op", op),
	)

	return c.JSON(&DeleteByIDResponse{
		Status:  fiber.StatusOK,
		Message: "entity has been deleted",
	})
}

type UpdateByIDRequest struct {
	Name        string `json:"name"`
	Surname     string `json:"surname"`
	Patronymic  string `json:"patronymic"`
	Age         int    `json:"age"`
	Gender      string `json:"gender"`
	Nationality string `json:"nationality"`
}

type UpdateByIDResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func (h handlers) UpdateByID(c *fiber.Ctx) error {
	const op = "effectivemobile.UpdateByID()"

	var body UpdateByIDRequest

	err := c.BodyParser(&body)
	if err != nil {
		h.log.Debug(
			"неправильно сформирован запрос",
			slog.String("source", source),
			slog.String("op", op),
			slog.Any("error", err),
		)

		return fiber.ErrBadRequest
	}

	id := c.Params("id")
	if id == "" {
		h.log.Debug(
			"отсутсвуют параметры",
			slog.String("source", source),
			slog.String("op", op),
			slog.String("error", "absent parameters"),
		)

		return fiber.ErrBadRequest
	}

	h.log.Debug(
		"получен запрос на обновление",
		slog.String("source", source),
		slog.String("op", op),
		slog.Any("parameters", []string{id}),
		slog.Any("body", body),
	)

	err = h.Service.UpdateByID(id, c.Body())
	if err != nil {
		switch {
		case errors.Is(err, effectivemobileservice.ErrStorageNotFound):
			return fiber.ErrNotFound

		case errors.Is(err, effectivemobileservice.ErrStorageBadRequest):
			return fiber.ErrBadRequest

		default:
			return fiber.ErrInternalServerError
		}

	}
	h.log.Debug(
		"обработан запрос на обновление",
		slog.String("source", source),
		slog.String("op", op),
	)

	return c.JSON(UpdateByIDResponse{
		Status:  fiber.StatusOK,
		Message: "entity has been updated",
	})
}

type CreateRequest struct {
	Name       string `json:"name"`
	Surname    string `json:"surname"`
	Patronymic string `json:"patronymic"`
}

type CreateResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func (h handlers) Create(c *fiber.Ctx) error {
	const op = "effectivemobile.UpdateByID()"

	var body CreateRequest

	err := c.BodyParser(&body)
	if err != nil {
		h.log.Debug(
			"неправильно сформирован запрос",
			slog.String("source", source),
			slog.String("op", op),
			slog.Any("error", err),
		)

		return fiber.ErrBadRequest
	}

	if body.Name == "" {
		h.log.Debug(
			"неправильно сформирован запрос",
			slog.String("source", source),
			slog.String("op", op),
			slog.Any("error", err),
		)

		return fiber.ErrBadRequest
	}

	if body.Surname == "" {
		h.log.Debug(
			"неправильно сформирован запрос",
			slog.String("source", source),
			slog.String("op", op),
			slog.Any("error", err),
		)

		return fiber.ErrBadRequest
	}

	h.log.Debug(
		"получен запрос на обновление",
		slog.String("source", source),
		slog.String("op", op),
		slog.Any("body", body),
	)

	err = h.Service.Create(body.Name, body.Surname, body.Patronymic)
	if err != nil {
		switch {
		case errors.Is(err, effectivemobileservice.ErrStorageNotFound):
			return fiber.ErrNotFound

		case errors.Is(err, effectivemobileservice.ErrClientNotFound):
			return fiber.ErrNotFound

		default:
			return fiber.ErrInternalServerError
		}
	}
	h.log.Debug(
		"обработан запрос на создание",
		slog.String("source", source),
		slog.String("op", op),
	)

	return c.JSON(UpdateByIDResponse{
		Status:  fiber.StatusOK,
		Message: "entity has been created",
	})
}

type SelectRequest struct {
	Limit  []int  `json:"limit"`
	Filter Filter `json:"filter"`
}

type SelectResponse struct {
	Status  int           `json:"status"`
	Message string        `json:"message"`
	Result  []storage.Row `json:"result"`
}

type Filter struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

var (
	FilterName        = "name"
	FilterSurname     = "surname"
	FilterPatronymic  = "patronymic"
	FilterAge         = "age"
	FilterGender      = "gender"
	FilterNationality = "nationality"
)

func (h handlers) Select(c *fiber.Ctx) error {
	const op = "effectivemobile.Select()"

	var body SelectRequest

	if c.Body() != nil {
		err := c.BodyParser(&body)
		if err != nil {
			h.log.Debug(
				"неправильно сформирован запрос",
				slog.String("source", source),
				slog.String("op", op),
				slog.Any("error", err),
			)

			return fiber.ErrBadRequest
		}
	}

	id := c.Params("id")
	if id != "" {
		body = SelectRequest{}
	}

	h.log.Debug(
		"получен запрос на получение",
		slog.String("source", source),
		slog.String("op", op),
		slog.Any("parameters", []string{id}),
		slog.Any("body", body),
	)

	if id == "" && c.Body() == nil {
		h.log.Debug(
			"неправильно сформирован запрос",
			slog.String("source", source),
			slog.String("op", op),
			slog.Any("error", "empty body and id parameter"),
		)

		return fiber.ErrBadRequest
	}

	if body.Limit == nil {
		body.Limit = []int{0, h.config.SelectLimit}
	}

	filters := []string{FilterName, FilterSurname, FilterPatronymic, FilterAge, FilterGender, FilterNationality}

	if body.Filter.Type == "" {
		if body.Filter.Value != "" {
			h.log.Debug(
				"неправильно сформирован запрос",
				slog.String("source", source),
				slog.String("op", op),
				slog.Any("error", "malfored body"),
			)

			return fiber.ErrBadRequest
		}
	} else {
		if body.Filter.Value == "" {
			h.log.Debug(
				"неправильно сформирован запрос",
				slog.String("source", source),
				slog.String("op", op),
				slog.Any("error", "malfored body"),
			)

			return fiber.ErrBadRequest
		}

		exists := false

		for _, filter := range filters {
			if body.Filter.Type == filter {
				exists = true
			}
		}

		if !exists {
			h.log.Debug(
				"неправильно сформирован запрос",
				slog.String("source", source),
				slog.String("op", op),
				slog.Any("error", "malfored body"),
			)

			return fiber.ErrBadRequest
		}
	}

	r, err := h.Service.Select(id, body.Limit, body.Filter.Type, body.Filter.Value)
	if err != nil {
		switch {
		case errors.Is(err, effectivemobileservice.ErrStorageNotFound):
			return fiber.ErrNotFound

		default:
			return fiber.ErrInternalServerError
		}
	}

	h.log.Debug(
		"обработан запрос на получение",
		slog.String("source", source),
		slog.String("op", op),
	)

	return c.JSON(SelectResponse{
		Status:  fiber.StatusOK,
		Message: "data found",
		Result:  r,
	})
}

// МОКИ

type UnimplementedHandlers struct{}

func (u UnimplementedHandlers) DeleteByID(c *fiber.Ctx) error {
	return nil
}

func (u UnimplementedHandlers) UpdateByID(c *fiber.Ctx) error {
	return nil
}

func (u UnimplementedHandlers) Create(c *fiber.Ctx) error {
	return nil
}

func (u UnimplementedHandlers) Select(c *fiber.Ctx) error {
	return nil
}
