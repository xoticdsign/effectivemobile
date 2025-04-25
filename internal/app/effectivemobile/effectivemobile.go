package effectivemobile

import (
	"errors"
	"fmt"
	"log/slog"
	"net"

	"github.com/gofiber/fiber/v2"
	effectivemobileservice "github.com/xoticdsign/effectivemobile/internal/service/effectivemobile"
	storage "github.com/xoticdsign/effectivemobile/internal/storage/postgresql"
	"github.com/xoticdsign/effectivemobile/internal/utils/config"
)

const source = "effectivemobile"

type App struct {
	Server Server

	log    *slog.Logger
	config config.EffectiveMobileConfig
}

type Handlerer interface {
	DeleteByID(c *fiber.Ctx) error
	UpdateByID(c *fiber.Ctx) error
	Create(c *fiber.Ctx) error
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

	emservice := effectivemobileservice.New(config, storage, log)

	h := handlers{
		Service: emservice.S.Handlers,

		log:    log,
		config: config,
	}

	f.Delete("/delete/:id", h.DeleteByID)
	f.Put("/update/:id", h.UpdateByID)
	f.Post("/create")

	return &App{
		Server: Server{
			Implementation: f,
			Handlers:       h,
		},

		log:    log,
		config: config,
	}
}

func (a *App) Run() error {
	const op = "effectivemobile.Run()"

	err := a.Server.Implementation.Listen(net.JoinHostPort(a.config.Host, a.config.Port))
	if err != nil {
		return fmt.Errorf("%s @ %v", op, err)
	}
	return nil
}

func (a *App) Shutdown() error {
	const op = "effectivemobile.Shutdown()"

	err := a.Server.Implementation.Shutdown()
	if err != nil {
		return fmt.Errorf("%s @ %v", op, err)
	}
	return nil
}

type Servicer interface {
	DeleteByID(id string) error
	UpdateByID(id string) error
	Create(name string, surname string, patronymic string) error
}

type handlers struct {
	UnimplementedHandlers

	Service Servicer

	log    *slog.Logger
	config config.EffectiveMobileConfig
}

func (h handlers) DeleteByID(c *fiber.Ctx) error {
	err := h.Service.DeleteByID(c.Params("id"))
	if err != nil {
		// ERROR HANDLING
	}
	return nil
}

func (h handlers) UpdateByID(c *fiber.Ctx) error {
	err := h.Service.UpdateByID(c.Params("id"))
	if err != nil {
		// ERROR HANDLING
	}
	return nil
}

type CreateRequest struct {
	Name       string `json:"name"`
	Surname    string `json:"surname"`
	Patronymic string `json:"patronymic"`
}

func (h handlers) Create(c *fiber.Ctx) error {
	var body CreateRequest

	err := c.BodyParser(&body)
	if err != nil {
		return fiber.ErrBadRequest
	}

	err = h.Service.Create(body.Name, body.Surname, body.Patronymic)
	if err != nil {
		// ERROR HANDLING
	}

	return nil
}

// МОКИ

type UnimplementedHandlers struct{}

func (u UnimplementedHandlers) DeleteByID(c *fiber.Ctx) error {
	return nil
}

func (h UnimplementedHandlers) UpdateByID(c *fiber.Ctx) error {
	return nil
}

func (h UnimplementedHandlers) Create(c *fiber.Ctx) error {
	return nil
}
