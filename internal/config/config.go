// Package config загружает параметры запуска сервиса из флагов командной строки
// и переменных окружения. Флаги имеют приоритет над переменными окружения.
package config

import (
	"flag"
	"os"
)

const (
	defaultRunAddress = "localhost:8080"
	defaultAuthSecret = "gophermart-dev-secret"
)

// Config содержит параметры запуска сервиса накопительной системы лояльности.
type Config struct {
	// RunAddress — адрес и порт HTTP-сервера.
	RunAddress string
	// DatabaseURI — строка подключения к PostgreSQL.
	DatabaseURI string
	// AccrualAddress — базовый URL системы расчёта начислений.
	AccrualAddress string
	// AuthSecret — секрет подписи JWT-токенов аутентификации.
	AuthSecret string
}

// Parse разбирает аргументы командной строки и переменные окружения
// RUN_ADDRESS, DATABASE_URI, ACCRUAL_SYSTEM_ADDRESS и AUTH_SECRET.
func Parse(args []string) (*Config, error) {
	cfg := &Config{}

	fs := flag.NewFlagSet("gophermart", flag.ContinueOnError)
	fs.StringVar(&cfg.RunAddress, "a", "", "адрес и порт запуска сервиса")
	fs.StringVar(&cfg.DatabaseURI, "d", "", "адрес подключения к базе данных")
	fs.StringVar(&cfg.AccrualAddress, "r", "", "адрес системы расчёта начислений")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if cfg.RunAddress == "" {
		cfg.RunAddress = os.Getenv("RUN_ADDRESS")
	}
	if cfg.DatabaseURI == "" {
		cfg.DatabaseURI = os.Getenv("DATABASE_URI")
	}
	if cfg.AccrualAddress == "" {
		cfg.AccrualAddress = os.Getenv("ACCRUAL_SYSTEM_ADDRESS")
	}
	if cfg.AuthSecret == "" {
		cfg.AuthSecret = os.Getenv("AUTH_SECRET")
	}

	if cfg.RunAddress == "" {
		cfg.RunAddress = defaultRunAddress
	}
	if cfg.AuthSecret == "" {
		cfg.AuthSecret = defaultAuthSecret
	}

	return cfg, nil
}
