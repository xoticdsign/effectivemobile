package logger

import (
	"fmt"
	"log/slog"
	"os"
)

func New(env string) (*slog.Logger, error) {
	const op = "logger.New()"

	var log *slog.Logger

	switch env {
	case "local":
		log = slog.New(slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		))

	case "dev":
		log = slog.New(slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		))

	case "prod":
		log = slog.New(slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		))

	default:
		return nil, fmt.Errorf("%s @ error: environment is unknown", op)
	}

	return log, nil
}
