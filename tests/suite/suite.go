package suite

import (
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

func New(t *testing.T) *Suite {
	t.Helper()
	t.Parallel()

	log := slog.New(logger.SilentHandler{})

	var config config.Config

	err := godotenv.Load("t.env")
	if err != nil {
		panic(err)
	}

	err = cleanenv.ReadEnv(&config)
	if err != nil {
		panic(err)
	}

	return &Suite{
		T: t,
		Log: &logger.Logger{
			Log: log,
		},
		Config: config,
	}
}
