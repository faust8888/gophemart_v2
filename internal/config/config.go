// Package config загружает параметры запуска сервиса из флагов командной строки
// и переменных окружения. Если значение задано и флагом, и переменной окружения,
// используется переменная окружения.
package config

import (
	"flag"
	"os"
)

const (
	// EnvRunAddress — переменная окружения с адресом и портом запуска HTTP-сервиса.
	EnvRunAddress = "RUN_ADDRESS"
	// EnvDatabaseURI — переменная окружения со строкой подключения к PostgreSQL.
	EnvDatabaseURI = "DATABASE_URI"
	// EnvAccrualAddress — переменная окружения с адресом системы расчёта начислений.
	EnvAccrualAddress = "ACCRUAL_SYSTEM_ADDRESS"

	defaultRunAddress = "localhost:8080"
	defaultAuthSecret = "gophermart-dev-secret"
)

// Config содержит параметры запуска сервиса накопительной системы лояльности.
type Config struct {
	// RunAddress — адрес и порт HTTP-сервера (RUN_ADDRESS / -a).
	RunAddress string
	// DatabaseURI — строка подключения к PostgreSQL (DATABASE_URI / -d).
	DatabaseURI string
	// AccrualAddress — базовый URL системы расчёта начислений (ACCRUAL_SYSTEM_ADDRESS / -r).
	AccrualAddress string
	// AuthSecret — секрет подписи JWT-токенов аутентификации.
	AuthSecret string
}

// Parse разбирает флаги командной строки и переменные окружения.
// Поддерживаются:
//   - адрес запуска: флаг -a или RUN_ADDRESS;
//   - база данных: флаг -d или DATABASE_URI;
//   - система начислений: флаг -r или ACCRUAL_SYSTEM_ADDRESS.
func Parse(args []string) (*Config, error) {
	cfg := &Config{
		RunAddress: defaultRunAddress,
	}

	fs := flag.NewFlagSet("gophermart", flag.ContinueOnError)
	fs.StringVar(&cfg.RunAddress, "a", defaultRunAddress, "адрес и порт запуска сервиса (env "+EnvRunAddress+")")
	fs.StringVar(&cfg.DatabaseURI, "d", "", "адрес подключения к базе данных (env "+EnvDatabaseURI+")")
	fs.StringVar(&cfg.AccrualAddress, "r", "", "адрес системы расчёта начислений (env "+EnvAccrualAddress+")")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	cfg.RunAddress = envOverride(cfg.RunAddress, EnvRunAddress)
	cfg.DatabaseURI = envOverride(cfg.DatabaseURI, EnvDatabaseURI)
	cfg.AccrualAddress = envOverride(cfg.AccrualAddress, EnvAccrualAddress)

	if cfg.AuthSecret = os.Getenv("AUTH_SECRET"); cfg.AuthSecret == "" {
		cfg.AuthSecret = defaultAuthSecret
	}

	return cfg, nil
}

func envOverride(flagValue, envName string) string {
	if envValue := os.Getenv(envName); envValue != "" {
		return envValue
	}
	return flagValue
}
