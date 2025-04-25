package effectivemobile

import (
	"errors"
	"fmt"
	"log/slog"
	"net"

	"github.com/gofiber/fiber/v2"
	effectivemobileservice "github.com/xoticdsign/effectivemobile/internal/service/effectivemobile"
	"github.com/xoticdsign/effectivemobile/internal/utils/config"
)

type App struct {
	Server Server

	log    *slog.Logger
	config config.EffectiveMobileConfig
}

type Handlerer interface{}

type Server struct {
	Implementation *fiber.App
	Handlers       Handlerer
}

func New(config config.EffectiveMobileConfig, log *slog.Logger) *App {
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

	return &App{
		Server: Server{
			Implementation: f,
			Handlers: handlers{
				Service: effectivemobileservice.New(config, log),

				log:    log,
				config: config,
			},
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

type Servicer interface{}

type handlers struct {
	UnimplementedHandlers

	Service Servicer

	log    *slog.Logger
	config config.EffectiveMobileConfig
}

// МОКИ

type UnimplementedHandlers struct{}
