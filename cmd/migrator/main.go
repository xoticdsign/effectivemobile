package main

import (
	"errors"
	"fmt"
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

	m, err := migrate.New("file://"+config.MigrationsPath, fmt.Sprintf("%s&&x-migrations-table=%s", config.Storage.PostgreSQL.Address, config.MigrationsTable))
	if err != nil {
		log.Log.Error(
			"невозможно инициализировать мигратор",
			slog.String("source", source),
			slog.String("op", op),
			slog.Any("error", err),
		)

		panic(err)
	}

	log.Log.Debug(
		"начинается миграция",
		slog.String("source", source),
		slog.String("op", op),
	)

	switch config.MigrationsDirection {
	case directionUp:
		log.Log.Debug(
			"миграция вверх",
			slog.String("source", source),
			slog.String("op", op),
		)

		err := m.Up()
		if err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				log.Log.Warn(
					"нечего мигрировать или отсутствуют файлы миграций",
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
		log.Log.Debug(
			"миграция вниз",
			slog.String("source", source),
			slog.String("op", op),
		)

		err := m.Down()
		if err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				log.Log.Warn(
					"нечего мигрировать или отсутствуют файлы миграций",
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
