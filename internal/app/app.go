package app

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	effectivemobileapp "github.com/xoticdsign/effectivemobile/internal/app/effectivemobile"
	"github.com/xoticdsign/effectivemobile/internal/lib/logger"
	storage "github.com/xoticdsign/effectivemobile/internal/storage/postgresql"
	"github.com/xoticdsign/effectivemobile/internal/utils/config"
)

const source = "app"

type App struct {
	EffectiveMobile *effectivemobileapp.App
	Storage         *storage.Storage

	log    *logger.Logger
	config config.Config
}

func New() (*App, error) {
	const op = "app.New()"

	config, err := config.New()
	if err != nil {
		return nil, fmt.Errorf("%s.%v", op, err)
	}

	log, err := logger.New(config.LogMode)
	if err != nil {
		return nil, fmt.Errorf("%s.%v", op, err)
	}

	log.Log.Debug(
		"инициализация хранилища",
		slog.String("source", source),
		slog.String("op", op),
		slog.Any("config", config.Storage.PostgreSQL),
	)

	storage, err := storage.New(config.Storage.PostgreSQL, log.Log)
	if err != nil {
		return nil, fmt.Errorf("%s.%v", op, err)
	}

	log.Log.Debug(
		"инициализация effectivemobile",
		slog.String("source", source),
		slog.String("op", op),
		slog.Any("config", config.EffectiveMobile),
	)

	emapp := effectivemobileapp.New(config.EffectiveMobile, storage, log.Log)

	return &App{
		EffectiveMobile: emapp,
		Storage:         storage,

		log:    log,
		config: config,
	}, nil
}

func (a *App) Run() {
	const op = "app.Run()"

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	errChan := make(chan error, 1)

	a.log.Log.Info(
		"запуск сервера",
		slog.String("source", source),
		slog.String("op", op),
	)

	go func() {
		err := a.EffectiveMobile.Run()
		if err != nil {
			errChan <- err
		}
	}()

	a.log.Log.Info(
		"сервер запущен",
		slog.String("source", source),
		slog.String("op", op),
	)

	select {
	case sig := <-sigChan:
		a.log.Log.Debug(
			"сервер получил стоп-сигнал",
			slog.String("source", source),
			slog.String("op", op),
			slog.Any("signal", sig),
		)

	case err := <-errChan:
		a.log.Log.Error(
			"во время работы произошла оишбка, сервер вынужден остановиться",
			slog.String("source", source),
			slog.String("op", op),
			slog.Any("error", err),
		)
	}

	a.log.Log.Info(
		"попытка выполнить gracefull shutdown",
		slog.String("source", source),
		slog.String("op", op),
	)

	err := a.shutdown()
	if err != nil {
		a.log.Log.Error(
			"не удалось выполнить gracefull shutdown, выполняется принудительная остановка",
			slog.String("source", source),
			slog.String("op", op),
			slog.Any("error", err),
		)
	}
	a.log.Log.Info(
		"выполнен gracefull shutdown",
		slog.String("source", source),
		slog.String("op", op),
	)
}

func (a *App) shutdown() error {
	const op = "app.shutdown()"

	a.log.Log.Debug(
		"gracefull shutdown для храналища",
		slog.String("source", source),
		slog.String("op", op),
	)

	err := a.Storage.Shutdown()
	if err != nil {
		return fmt.Errorf("%s.%v", op, err)
	}

	a.log.Log.Debug(
		"shutdown для логов",
		slog.String("source", source),
		slog.String("op", op),
	)

	err = a.log.Shutdown()
	if err != nil {
		return fmt.Errorf("%s.%v", op, err)
	}

	a.log.Log.Debug(
		"gracefull shutdown для effectivemobile",
		slog.String("source", source),
		slog.String("op", op),
	)

	err = a.EffectiveMobile.Shutdown()
	if err != nil {
		return fmt.Errorf("%s.%v", op, err)
	}

	return nil
}
