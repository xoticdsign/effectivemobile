package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/xoticdsign/effectivemobile/internal/utils"
)

type Logger struct {
	Log  *slog.Logger
	File *os.File
}

type silentHandler struct{}

func (s silentHandler) Enabled(_ context.Context, _ slog.Level) bool  { return false }
func (s silentHandler) Handle(_ context.Context, _ slog.Record) error { return nil }
func (s silentHandler) WithAttrs(_ []slog.Attr) slog.Handler          { return s }
func (s silentHandler) WithGroup(_ string) slog.Handler               { return s }

func New(logMode string) (*Logger, error) {
	const op = "logger.New()"

	var log *slog.Logger
	var file *os.File
	var err error

	switch logMode {
	case "silent":
		log = slog.New(silentHandler{})

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
		Log:  log,
		File: file,
	}, nil
}

func (l *Logger) Shutdown() {
	if l.File != nil {
		l.File.Close()
	}
}
