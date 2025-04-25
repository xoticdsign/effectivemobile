package effectivemobile

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type App struct {
	Server Server

	// log *slog.Logger
	// config utils.Config
}

type Handlerer interface{}

type Server struct {
	Implementation *fiber.App
	Handlers       Handlerer
}

func New( /* config utils.Config, log *slog.Logger */ ) App {
	f := fiber.New(fiber.Config{
		// ReadTimeout: ,
		// WriteTimeout: ,
		// IdleTimeout: ,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			var e *fiber.Error

			if errors.As(err, &e) {
				return c.JSON(e)
			}
			return c.JSON(fiber.ErrInternalServerError)
		},
		AppName: "effectivemobile",
	})

	return App{
		Server: Server{
			Implementation: f,
			Handlers:       handlers{ /* Service: effectivemobileservice.New() */ },
		},
	}
}

func (a App) Run() error {
	const op = "effectivemobile.Run()"

	err := a.Server.Implementation.Listen( /* config. */ )
	if err != nil {
		return fmt.Errorf("%s @ %v", op, err)
	}
	return nil
}

func (a App) Shutdown() error {
	const op = "effectivemobile.Shutdown()"

	err := a.Server.Implementation.Shutdown()
	if err != nil {
		return fmt.Errorf("%s @ %v", op, err)
	}
	return nil
}

type handlers struct {
	UnimplementedHandlers

	// Service effectivemobileservice.Service
}

type UnimplementedHandlers struct{}
