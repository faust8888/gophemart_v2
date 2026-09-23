// Package config загружает параметры запуска сервиса из флагов командной строки
// и переменных окружения. Если значение задано и флагом, и переменной окружения,
// используется переменная окружения.
package config

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
)

const (
	// EnvRunAddress — переменная окружения с адресом и портом запуска HTTP-сервиса.
	EnvRunAddress = "RUN_ADDRESS"
	// EnvDatabaseURI — переменная окружения со строкой подключения к PostgreSQL.
	EnvDatabaseURI = "DATABASE_URI"
	// EnvAccrualAddress — переменная окружения с адресом системы расчёта начислений.
	EnvAccrualAddress = "ACCRUAL_SYSTEM_ADDRESS"
	// EnvAuthSecret — переменная окружения с секретом подписи JWT.
	EnvAuthSecret = "AUTH_SECRET"
	// EnvEnvironment — переменная окружения с именем окружения (prod/production запрещает пустой AUTH_SECRET).
	EnvEnvironment = "GOPHERMART_ENV"

	defaultRunAddress    = "localhost:8080"
	ephemeralSecretBytes = 32
)

// ErrAuthSecretRequired возвращается, если в production не задан AUTH_SECRET.
var ErrAuthSecretRequired = errors.New("AUTH_SECRET is required in production")

// Config содержит параметры запуска сервиса накопительной системы лояльности.
type Config struct {
	// RunAddress — адрес и порт HTTP-сервера (RUN_ADDRESS / -a).
	RunAddress string
	// DatabaseURI — строка подключения к PostgreSQL (DATABASE_URI / -d).
	DatabaseURI string
	// AccrualAddress — базовый URL системы расчёта начислений (ACCRUAL_SYSTEM_ADDRESS / -r).
	AccrualAddress string
	// AuthSecret — секрет подписи JWT-токенов аутентификации (AUTH_SECRET).
	AuthSecret string
	// EphemeralAuthSecret — признак, что секрет сгенерирован на время процесса и не задан явно.
	EphemeralAuthSecret bool
}

// Parse разбирает флаги командной строки и переменные окружения.
// Поддерживаются:
//   - адрес запуска: флаг -a или RUN_ADDRESS;
//   - база данных: флаг -d или DATABASE_URI;
//   - система начислений: флаг -r или ACCRUAL_SYSTEM_ADDRESS;
//   - секрет JWT: AUTH_SECRET (в production обязателен).
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
	cfg.AuthSecret = os.Getenv(EnvAuthSecret)

	if err := ResolveAuthSecret(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// ResolveAuthSecret проверяет секрет JWT: в production пустой AUTH_SECRET недопустим,
// иначе генерируется одноразовый секрет текущего процесса.
func ResolveAuthSecret(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	if cfg.AuthSecret != "" {
		cfg.EphemeralAuthSecret = false
		return nil
	}
	if IsProduction() {
		return ErrAuthSecretRequired
	}
	secret, err := newEphemeralAuthSecret()
	if err != nil {
		return err
	}
	cfg.AuthSecret = secret
	cfg.EphemeralAuthSecret = true
	return nil
}

// IsProduction сообщает, запущено ли приложение в production-окружении.
func IsProduction() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(EnvEnvironment))) {
	case "prod", "production":
		return true
	default:
		return false
	}
}

func envOverride(flagValue, envName string) string {
	if envValue := os.Getenv(envName); envValue != "" {
		return envValue
	}
	return flagValue
}

func newEphemeralAuthSecret() (string, error) {
	buf := make([]byte, ephemeralSecretBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate auth secret: %w", err)
	}
	return hex.EncodeToString(buf), nil
}
