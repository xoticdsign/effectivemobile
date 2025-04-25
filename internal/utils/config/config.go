package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	LogMode string `env:"LOG_MODE env-required"`

	EffectiveMobile EffectiveMobileConfig
	Storage         StorageConfig
}

type EffectiveMobileConfig struct {
	Host         string        `env:"SERVER_HOST env-required"`
	Port         string        `env:"SERVER_PORT env-required"`
	ReadTimeout  time.Duration `env:"SERVER_READTIMEOUT env-required"`
	WriteTimeout time.Duration `env:"SERVER_WRITETIMEOUT env-required"`
	IdleTimeout  time.Duration `env:"SERVER_IDLETIMEOUT env-required"`
}

type StorageConfig struct {
	PostgreSQL PostgreSQLConfig
}

type PostgreSQLConfig struct {
	// TODO
}

func New() (Config, error) {
	const op = "config.New()"

	var config Config

	err := cleanenv.ReadConfig(".env", &config)
	if err != nil {
		return Config{}, fmt.Errorf("%s @ %v", op, err)
	}
	return config, nil
}
