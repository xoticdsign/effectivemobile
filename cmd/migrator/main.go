package main

import (
	"errors"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/xoticdsign/effectivemobile/internal/lib/logger"
	"github.com/xoticdsign/effectivemobile/internal/utils/config"
)

const source = "migrator"

const (
	directionUp   = "up"
	directionDown = "down"
)

func main() {
	const op = "main()"

	config, err := config.New()
	if err != nil {
		panic(err)
	}

	log, err := logger.New(config.LogMode)
	if err != nil {
		panic(err)
	}

	m, err := migrate.New("file://"+config.MigrationsPath, config.Storage.PostgreSQL.Address)
	if err != nil {
		log.Log.Error(
			"невозможно инициализировать мигратор",
			slog.String("source", source),
			slog.String("op", op),
			slog.Any("error", err),
		)

		panic(err)
	}

	log.Log.Info(
		"начинается миграция",
		slog.String("source", source),
		slog.String("op", op),
	)

	switch config.MigrationsDirection {
	case directionUp:
		err := m.Up()
		if err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				log.Log.Info(
					"нечего мигрировать",
					slog.String("source", source),
					slog.String("op", op),
				)
			} else {
				log.Log.Error(
					"невозможно мигрировать",
					slog.String("source", source),
					slog.String("op", op),
					slog.Any("error", err),
				)

				panic(err)
			}
		}

	case directionDown:
		err := m.Down()
		if err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				log.Log.Info(
					"нечего мигрировать",
					slog.String("source", source),
					slog.String("op", op),
				)
			} else {
				log.Log.Error(
					"невозможно мигрировать",
					slog.String("source", source),
					slog.String("op", op),
					slog.Any("error", err),
				)

				panic(err)
			}
		}

	default:
		log.Log.Error(
			"неверное направление миграции",
			slog.String("source", source),
			slog.String("op", op),
		)

		panic("неверное направление миграции")
	}

	log.Log.Info(
		"миграция завершена",
		slog.String("source", source),
		slog.String("op", op),
	)
}
