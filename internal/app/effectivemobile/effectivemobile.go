package effectivemobile

import (
	"errors"
	"fmt"
	"log/slog"
	"net"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/swagger"

	_ "github.com/xoticdsign/effectivemobile/docs"
	"github.com/xoticdsign/effectivemobile/internal/client"
	effectivemobileservice "github.com/xoticdsign/effectivemobile/internal/service/effectivemobile"
	storage "github.com/xoticdsign/effectivemobile/internal/storage/postgresql"
	"github.com/xoticdsign/effectivemobile/internal/utils/config"
)

const source = "effectivemobile"

var (
	DeleteByIDHanlder    = "delete"
	DeleteByIDParameters = ":id"
	UpdateByIDHandler    = "update"
	UpdateByIDParameters = ":id"
	CreateHandler        = "create"
	SelectHandler        = "select"
	SelectParameters     = ":id?"
)

var (
	DeleteByIDSuccess = "entity has been deleted"
	UpdateByIDSuccess = "entity has been updated"
	CreateSuccess     = "entity has been created"
	SelectSuccess     = "entity(ies) found"
)

type App struct {
	Server Server
	Client *client.Client
	Log    *slog.Logger
	Config config.EffectiveMobileConfig
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

	h := Handlers{
		Service: emservice.S.Handlers,
		Log:     log,
		Config:  config,
	}

	f.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	f.Delete(fmt.Sprintf("/%s/%s", DeleteByIDHanlder, DeleteByIDParameters), h.DeleteByID)
	f.Put(fmt.Sprintf("/%s/%s", UpdateByIDHandler, UpdateByIDParameters), h.UpdateByID)
	f.Post(fmt.Sprintf("/%s", CreateHandler), h.Create)
	f.Get(fmt.Sprintf("/%s/%s", SelectHandler, SelectParameters), h.Select)
	f.Get("/swagger/*", swagger.New(swagger.ConfigDefault))

	return &App{
		Server: Server{
			Implementation: f,
			Handlers:       h,
		},
		Client: client,
		Log:    log,
		Config: config,
	}
}

func (a *App) Run() error {
	const op = "effectivemobile.Run()"

	err := a.Server.Implementation.Listen(net.JoinHostPort(a.Config.Host, a.Config.Port))
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

type Handlers struct {
	UnimplementedHandlers

	Service Servicer
	Log     *slog.Logger
	Config  config.EffectiveMobileConfig
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type DeleteByIDResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// @description Удаляет запись из базы данных по заданному идентификатору.
//
// @id          delete
// @tags        Операции
//
// @summary     Удаление записи по ID
// @produce     json
// @param       id  path     string             true "Идентификатор записи"
// @success     200 {object} DeleteByIDResponse "Возвращается, если удаление прошло успешно"
// @failure     400 {object} ErrorResponse      "Возвращается, если запрос был сформирован неправильно"
// @failure     404 {object} ErrorResponse      "Возвращается, если запрашиваемая запись не была найдена"
// @failure     405 {object} ErrorResponse      "Возвращается, если был использован неправильный метод"
// @failure     500 {object} ErrorResponse      "Возвращается, если во время работы хранилища произошла ошибка"
// @router      /delete/{id} [delete]
func (h Handlers) DeleteByID(c *fiber.Ctx) error {
	const op = "effectivemobile.DeleteByID()"

	id := c.Params("id")
	if id == "" {
		h.Log.Debug(
			"отсутсвуют параметры",
			slog.String("source", source),
			slog.String("op", op),
			slog.String("error", "absent parameters"),
		)

		return fiber.ErrBadRequest
	}

	h.Log.Debug(
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
	h.Log.Debug(
		"обработан запрос на удаление",
		slog.String("source", source),
		slog.String("op", op),
	)

	return c.JSON(&DeleteByIDResponse{
		Code:    fiber.StatusOK,
		Message: DeleteByIDSuccess,
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
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// @description Обновляет существующую записи в базе данных по ID при помощи данных получаемых в теле запроса.
//
// @id          update
// @tags        Операции
//
// @summary     Обновление записи по ID
// @produce     json
// @param       id   path     string             true "Идентификатор записи"
// @param       body body     UpdateByIDRequest  true "Тело запроса"
// @success     200  {object} UpdateByIDResponse "Возвращается, если обновление прошло успешно"
// @failure     400  {object} ErrorResponse      "Возвращается, если запрос был сформирован неправильно"
// @failure     404  {object} ErrorResponse      "Возвращается, если запрашиваемая запись не была найдена"
// @failure     405  {object} ErrorResponse      "Возвращается, если был использован неправильный метод"
// @failure     409  {object} ErrorResponse      "Возвращается, если переданные данные ничем не отличаются от уже существующих"
// @failure     500  {object} ErrorResponse      "Возвращается, если во время работы хранилища произошла ошибка"
// @router      /update/{id} [put]
func (h Handlers) UpdateByID(c *fiber.Ctx) error {
	const op = "effectivemobile.UpdateByID()"

	var body UpdateByIDRequest

	err := c.BodyParser(&body)
	if err != nil {
		h.Log.Debug(
			"неправильно сформирован запрос",
			slog.String("source", source),
			slog.String("op", op),
			slog.Any("error", err),
		)

		return fiber.ErrBadRequest
	}

	id := c.Params("id")
	if id == "" {
		h.Log.Debug(
			"отсутсвуют параметры",
			slog.String("source", source),
			slog.String("op", op),
			slog.String("error", "absent parameters"),
		)

		return fiber.ErrBadRequest
	}

	h.Log.Debug(
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

		case errors.Is(err, effectivemobileservice.ErrStorageConflict):
			return fiber.ErrConflict

		default:
			return fiber.ErrInternalServerError
		}

	}
	h.Log.Debug(
		"обработан запрос на обновление",
		slog.String("source", source),
		slog.String("op", op),
	)

	return c.JSON(&UpdateByIDResponse{
		Code:    fiber.StatusOK,
		Message: UpdateByIDSuccess,
	})
}

type CreateRequest struct {
	Name       string `json:"name"`
	Surname    string `json:"surname"`
	Patronymic string `json:"patronymic"`
}

type CreateResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// @description Создает новую запись с автозаполнением возраста, пола и национальности при помощи открытых API.
//
// @id          create
// @tags        Операции
//
// @summary     Создание записи
// @produce     json
// @param       body body     CreateRequest  true "Тело запроса"
// @success     200  {object} CreateResponse "Возвращается, если создание прошло успешно"
// @failure     400    {object} ErrorResponse  "Возвращается, если запрос был сформирован неправильно"
// @failure     404  {object} ErrorResponse  "Возвращается, если запрашиваемая запись не была найдена/во внешних API нет данных"
// @failure     405    {object} ErrorResponse  "Возвращается, если был использован неправильный метод"
// @failure     500  {object} ErrorResponse  "Возвращается, если во время работы хранилища/клиента произошла ошибка"
// @router      /create [post]
func (h Handlers) Create(c *fiber.Ctx) error {
	const op = "effectivemobile.UpdateByID()"

	var body CreateRequest

	err := c.BodyParser(&body)
	if err != nil {
		h.Log.Debug(
			"неправильно сформирован запрос",
			slog.String("source", source),
			slog.String("op", op),
			slog.Any("error", err),
		)

		return fiber.ErrBadRequest
	}

	if body.Name == "" {
		h.Log.Debug(
			"неправильно сформирован запрос",
			slog.String("source", source),
			slog.String("op", op),
			slog.Any("error", err),
		)

		return fiber.ErrBadRequest
	}

	if body.Surname == "" {
		h.Log.Debug(
			"неправильно сформирован запрос",
			slog.String("source", source),
			slog.String("op", op),
			slog.Any("error", err),
		)

		return fiber.ErrBadRequest
	}

	h.Log.Debug(
		"получен запрос на создание",
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
	h.Log.Debug(
		"обработан запрос на создание",
		slog.String("source", source),
		slog.String("op", op),
	)

	return c.JSON(&CreateResponse{
		Code:    fiber.StatusOK,
		Message: CreateSuccess,
	})
}

var (
	FilterName        = "name"
	FilterSurname     = "surname"
	FilterPatronymic  = "patronymic"
	FilterAge         = "age"
	FilterGender      = "gender"
	FilterNationality = "nationality"
)

type SelectResponse struct {
	Code    int           `json:"code"`
	Message string        `json:"message"`
	Result  []storage.Row `json:"result"`
}

// @description Возвращает запись/список записей с возможностью фильтрации и пагинации.
//
// @id          select
// @tags        Операции
//
// @summary     Получение записи(ей)
// @produce     json
// @param       id     query    string         false "Идентификатор записи"
// @param       filter query    string         false "Тип фильтра (name, surname, etc.)"
// @param       value  query    string         false "Значение фильтра"
// @param       start  query    int            false "Начальная позиция"
// @param       end    query    int            false "Конечная позиция"
// @success     200    {object} SelectResponse "Возвращается, если получение прошло успешно"
// @failure     400  {object} ErrorResponse  "Возвращается, если запрос был сформирован неправильно"
// @failure     404    {object} ErrorResponse  "Возвращается, если запрашиваемая запись(и) не была найдена"
// @failure     405  {object} ErrorResponse  "Возвращается, если был использован неправильный метод"
// @failure     500    {object} ErrorResponse  "Возвращается, если во время работы хранилища произошла ошибка"
// @router      /select [get]
func (h Handlers) Select(c *fiber.Ctx) error {
	const op = "effectivemobile.Select()"

	var id string

	id = c.Params("id")
	if id == "" {
		id = c.Query("id")
	}

	filter := c.Query("filter")
	value := c.Query("value")
	start := c.QueryInt("start", 0)
	end := c.QueryInt("end", h.Config.SelectLimit)

	if start >= end {
		h.Log.Debug(
			"неправильно сформирован запрос",
			slog.String("source", source),
			slog.String("op", op),
			slog.Any("error", "start parameter can't be greater than the end"),
		)

		return fiber.ErrBadRequest
	}

	if id == "" && (filter == "" && value == "") {
		h.Log.Debug(
			"неправильно сформирован запрос",
			slog.String("source", source),
			slog.String("op", op),
			slog.Any("error", "both filter and id parameters are empty"),
		)

		return fiber.ErrBadRequest
	}

	if id != "" && (filter != "" && value != "") {
		h.Log.Debug(
			"неправильно сформирован запрос",
			slog.String("source", source),
			slog.String("op", op),
			slog.Any("error", "filter and id parameters can't be used at the same time"),
		)

		return fiber.ErrBadRequest
	}

	if (filter != "" && value == "") || (filter == "" && value != "") {
		h.Log.Debug(
			"неправильно сформирован запрос",
			slog.String("source", source),
			slog.String("op", op),
			slog.Any("error", "filter incomplete"),
		)
		return fiber.ErrBadRequest
	}

	filters := map[string]bool{
		FilterName: true, FilterSurname: true, FilterPatronymic: true,
		FilterAge: true, FilterGender: true, FilterNationality: true,
	}

	if filter != "" && !filters[filter] {
		h.Log.Debug(
			"неправильно сформирован запрос",
			slog.String("source", source),
			slog.String("op", op),
			slog.Any("error", "unknown filter"),
		)
		return fiber.ErrBadRequest
	}

	h.Log.Debug(
		"получен запрос на получение",
		slog.String("source", source),
		slog.String("op", op),
		slog.Any("parameters", []interface{}{id, filter, value, start, end}),
	)

	r, err := h.Service.Select(id, []int{start, end}, filter, value)
	if err != nil {
		switch {
		case errors.Is(err, effectivemobileservice.ErrStorageNotFound):
			return fiber.ErrNotFound

		default:
			return fiber.ErrInternalServerError
		}
	}

	h.Log.Debug(
		"обработан запрос на получение",
		slog.String("source", source),
		slog.String("op", op),
	)

	return c.JSON(&SelectResponse{
		Code:    fiber.StatusOK,
		Message: SelectSuccess,
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
