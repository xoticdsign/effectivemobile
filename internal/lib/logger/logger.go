package logger

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/xoticdsign/effectivemobile/internal/utils"
)

type Logger struct {
	Log *slog.Logger

	file *os.File
}

func New(logMode string) (*Logger, error) {
	const op = "logger.New()"

	var log *slog.Logger
	var file *os.File
	var err error

	switch logMode {
	case "local":
		log = slog.New(slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		))

	case "dev":
		file, err = utils.GetLogFile("log/dev/dev.log.json")
		if err != nil {
			return nil, fmt.Errorf("%s.%v", op, err)
		}

		log = slog.New(slog.NewJSONHandler(
			file,
			&slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		))

	case "prod":
		file, err = utils.GetLogFile("log/log.json")
		if err != nil {
			return nil, fmt.Errorf("%s.%v", op, err)
		}

		log = slog.New(slog.NewJSONHandler(
			file,
			&slog.HandlerOptions{
				Level: slog.LevelInfo,
			},
		))

	default:
		return nil, fmt.Errorf("%s @ error: logModeironment is unknown", op)
	}

	return &Logger{
		Log: log,

		file: file,
	}, nil
}

func (l *Logger) Shutdown() {
	if l.file != nil {
		l.file.Close()
	}
}
