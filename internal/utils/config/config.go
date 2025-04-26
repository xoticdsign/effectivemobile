package config

import (
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	MigrationsPath      string `env:"MIGRATIONS_PATH" env-required:"true" env-description:"Путь до миграций"`
	MigrationsDirection string `env:"MIGRATIONS_DIRECTION" env-required:"true" env-description:"Направление миграций"`
	MigrationsTable     string `env:"MIGRATIONS_TABLE" env-required:"true" env-description:"Таблица миграций"`

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

	Client ClientConfig
}

type ClientConfig struct {
	Timeout time.Duration `env:"CLIENT_TIMEOUT" env-required:"true" env-description:"Таймаут клиента"`
}

type StorageConfig struct {
	PostgreSQL PostgreSQLConfig
}

type PostgreSQLConfig struct {
	Username string `env:"POSTGRESQL_USERNAME" env-required:"true" env-description:"Имя пользователя PostgreSQL"`
	Password string `env:"POSTGRESQL_PASSWORD" env-description:"Пароль PostgreSQL"`
	Host     string `env:"POSTGRESQL_HOST" env-required:"true" env-description:"Имя хоста PostgreSQL"`
	Port     string `env:"POSTGRESQL_PORT" env-required:"true" env-description:"Порт PostgreSQL"`
	Database string `env:"POSTGRESQL_DBNAME" env-required:"true" env-description:"БД PostgreSQL"`
	Table    string `env:"POSTGRESQL_TABLE" env-required:"true" env-description:"Таблица PostgreSQL"`
	SSL      string `env:"POSTGRESQL_SSLMODE" env-required:"true" env-description:"Режим SSL PostgreSQL"`
	Extra    string `env:"POSTGRESQL_EXTRA" env-description:"Дополнительные опции PostgreSQL"`
}

func New() (Config, error) {
	const op = "config.New()"

	var config Config

	err := godotenv.Load()
	if err != nil {
		return Config{}, err
	}

	err = cleanenv.ReadEnv(&config)
	if err != nil {
		fmt.Println("")

		header := "ПЕРЕМЕННЫЕ ОКРУЖЕНИЯ:"
		cleanenv.FUsage(os.Stdout, &config, &header)()

		fmt.Println("")

		return Config{}, err
	}

	return config, nil
}
