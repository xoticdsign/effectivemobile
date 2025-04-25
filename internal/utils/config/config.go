package config

import (
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	LogMode string `env:"LOG_MODE" env-required:"true" env-description:"Режим логгирования (local, dev, prod)"`

	EffectiveMobile EffectiveMobileConfig
	Storage         StorageConfig
}

type EffectiveMobileConfig struct {
	Host         string        `env:"SERVER_HOST" env-required:"true" env-description:"Имя хоста"`
	Port         string        `env:"SERVER_PORT" env-required:"true" env-description:"Порт сервера"`
	ReadTimeout  time.Duration `env:"SERVER_READTIMEOUT" env-required:"true" env-description:"Таймаут сервера на Read"`
	WriteTimeout time.Duration `env:"SERVER_WRITETIMEOUT" env-required:"true" env-description:"Таймаут сервера на Write"`
	IdleTimeout  time.Duration `env:"SERVER_IDLETIMEOUT" env-required:"true" env-description:"Таймаут сервера на Idle"`
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
		fmt.Println("")

		header := "ПЕРЕМЕННЫЕ ОКРУЖЕНИЯ:"
		cleanenv.FUsage(os.Stdout, &config, &header)()

		fmt.Println("")

		return Config{}, fmt.Errorf("%s @ %v", op, err)
	}

	return config, nil
}
