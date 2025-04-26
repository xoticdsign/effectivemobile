package suite

import (
	"context"
	"log/slog"
	"testing"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"github.com/xoticdsign/effectivemobile/internal/lib/logger"
	"github.com/xoticdsign/effectivemobile/internal/utils/config"
)

type Suite struct {
	T      *testing.T
	Log    *logger.Logger
	Config config.Config
}

type silentHandler struct{}

func (s silentHandler) Enabled(_ context.Context, _ slog.Level) bool  { return false }
func (s silentHandler) Handle(_ context.Context, _ slog.Record) error { return nil }
func (s silentHandler) WithAttrs(_ []slog.Attr) slog.Handler          { return s }
func (s silentHandler) WithGroup(_ string) slog.Handler               { return s }

func New(t *testing.T) (*Suite, error) {
	t.Helper()
	t.Parallel()

	log := slog.New(silentHandler{})

	var config config.Config

	err := godotenv.Load("t.env")
	if err != nil {
		return nil, err
	}

	err = cleanenv.ReadEnv(&config)
	if err != nil {
		return nil, err
	}

	return &Suite{
		T: t,
		Log: &logger.Logger{
			Log: log,
		},
	}, nil
}
