package logger

import (
	"context"
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
			return nil, err
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
			return nil, err
		}

		log = slog.New(slog.NewJSONHandler(
			file,
			&slog.HandlerOptions{
				Level: slog.LevelInfo,
			},
		))

	default:
		return nil, fmt.Errorf("log mode is unknown")
	}

	return &Logger{
		Log: log,

		file: file,
	}, nil
}

func (l *Logger) Shutdown() error {
	const op = "logger.Shutdown()"

	if l.file != nil {
		err := l.file.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// МОКИ

type SilentHandler struct{}

func (s SilentHandler) Enabled(_ context.Context, _ slog.Level) bool  { return false }
func (s SilentHandler) Handle(_ context.Context, _ slog.Record) error { return nil }
func (s SilentHandler) WithAttrs(_ []slog.Attr) slog.Handler          { return s }
func (s SilentHandler) WithGroup(_ string) slog.Handler               { return s }
