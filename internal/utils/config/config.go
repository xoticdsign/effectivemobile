package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	LogMode string `env:"LOG_MODE" env-required:"true"`

	EffectiveMobile EffectiveMobileConfig
	Storage         StorageConfig
}

type EffectiveMobileConfig struct {
	Host         string        `env:"SERVER_HOST" env-required:"true"`
	Port         string        `env:"SERVER_PORT" env-required:"true"`
	ReadTimeout  time.Duration `env:"SERVER_READTIMEOUT" env-required:"true"`
	WriteTimeout time.Duration `env:"SERVER_WRITETIMEOUT" env-required:"true"`
	IdleTimeout  time.Duration `env:"SERVER_IDLETIMEOUT" env-required:"true"`
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

	err := godotenv.Load()
	if err != nil {
		return Config{}, fmt.Errorf("%s @ %v", op, err)
	}

	err = cleanenv.ReadEnv(&config)
	if err != nil {
		return Config{}, fmt.Errorf("%s @ %v", op, err)
	}

	return config, nil
}
