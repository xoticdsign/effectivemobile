package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	LogMode string `env:"LOG_MODE env-required"`

	Host        string        `env:"SERVER_HOST env-required"`
	Port        string        `env:"SERVER_PORT env-required"`
	ReadTimeout time.Duration `env:"SERVER_READTIMEOUT env-required"`
	WriteTemout time.Duration `env:"SERVER_WRITETIMEOUT env-required"`
	IdleTime    time.Duration `env:"SERVER_IDLETIMEOUT env-required"`

	Storage Storage
}

type Storage struct {
	PostgreSQL PostgreSQL
}

type PostgreSQL struct {
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
