package config

import (
	"errors"
	"flag"
	"testing"
)

func TestParse_Defaults(t *testing.T) {
	t.Setenv(EnvRunAddress, "")
	t.Setenv(EnvDatabaseURI, "")
	t.Setenv(EnvAccrualAddress, "")
	t.Setenv(EnvAuthSecret, "")
	t.Setenv(EnvEnvironment, "")

	cfg, err := Parse(nil)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if cfg.RunAddress != defaultRunAddress {
		t.Errorf("RunAddress = %q, want %q", cfg.RunAddress, defaultRunAddress)
	}
	if cfg.DatabaseURI != "" {
		t.Errorf("DatabaseURI = %q, want empty", cfg.DatabaseURI)
	}
	if cfg.AccrualAddress != "" {
		t.Errorf("AccrualAddress = %q, want empty", cfg.AccrualAddress)
	}
	if cfg.AuthSecret == "" {
		t.Fatal("AuthSecret is empty, want ephemeral secret")
	}
	if !cfg.EphemeralAuthSecret {
		t.Fatal("EphemeralAuthSecret = false, want true when AUTH_SECRET is unset")
	}
}

func TestParse_Flags(t *testing.T) {
	t.Setenv(EnvRunAddress, "")
	t.Setenv(EnvDatabaseURI, "")
	t.Setenv(EnvAccrualAddress, "")
	t.Setenv(EnvAuthSecret, "")
	t.Setenv(EnvEnvironment, "")

	cfg, err := Parse([]string{
		"-a", "localhost:8081",
		"-d", "postgres://user:pass@localhost/gophermart?sslmode=disable",
		"-r", "http://localhost:8080",
	})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if cfg.RunAddress != "localhost:8081" {
		t.Errorf("RunAddress = %q, want flag value", cfg.RunAddress)
	}
	if cfg.DatabaseURI != "postgres://user:pass@localhost/gophermart?sslmode=disable" {
		t.Errorf("DatabaseURI = %q, want flag value", cfg.DatabaseURI)
	}
	if cfg.AccrualAddress != "http://localhost:8080" {
		t.Errorf("AccrualAddress = %q, want flag value", cfg.AccrualAddress)
	}
}

func TestParse_Env(t *testing.T) {
	t.Setenv(EnvRunAddress, "localhost:9090")
	t.Setenv(EnvDatabaseURI, "postgres://user:pass@localhost/db")
	t.Setenv(EnvAccrualAddress, "http://accrual:8080")
	t.Setenv(EnvAuthSecret, "from-env")
	t.Setenv(EnvEnvironment, "")

	cfg, err := Parse(nil)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if cfg.RunAddress != "localhost:9090" {
		t.Errorf("RunAddress = %q, want env value", cfg.RunAddress)
	}
	if cfg.DatabaseURI != "postgres://user:pass@localhost/db" {
		t.Errorf("DatabaseURI = %q, want env value", cfg.DatabaseURI)
	}
	if cfg.AccrualAddress != "http://accrual:8080" {
		t.Errorf("AccrualAddress = %q, want env value", cfg.AccrualAddress)
	}
	if cfg.AuthSecret != "from-env" {
		t.Errorf("AuthSecret = %q, want env value", cfg.AuthSecret)
	}
	if cfg.EphemeralAuthSecret {
		t.Fatal("EphemeralAuthSecret = true, want false when AUTH_SECRET is set")
	}
}

func TestParse_EnvOverridesFlags(t *testing.T) {
	t.Setenv(EnvRunAddress, "env:9090")
	t.Setenv(EnvDatabaseURI, "postgres://env")
	t.Setenv(EnvAccrualAddress, "http://env")

	cfg, err := Parse([]string{
		"-a", "flag:8081",
		"-d", "postgres://flag",
		"-r", "http://flag",
	})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if cfg.RunAddress != "env:9090" {
		t.Errorf("RunAddress = %q, want env value", cfg.RunAddress)
	}
	if cfg.DatabaseURI != "postgres://env" {
		t.Errorf("DatabaseURI = %q, want env value", cfg.DatabaseURI)
	}
	if cfg.AccrualAddress != "http://env" {
		t.Errorf("AccrualAddress = %q, want env value", cfg.AccrualAddress)
	}
}

func TestParse_UnknownFlag(t *testing.T) {
	_, err := Parse([]string{"-unknown"})
	if err == nil {
		t.Fatal("Parse() error = nil, want error")
	}
}

func TestParse_Help(t *testing.T) {
	_, err := Parse([]string{"-h"})
	if !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("Parse(-h) error = %v, want flag.ErrHelp", err)
	}
}

func TestParse_ProductionRequiresAuthSecret(t *testing.T) {
	t.Setenv(EnvAuthSecret, "")
	t.Setenv(EnvEnvironment, "production")

	_, err := Parse(nil)
	if !errors.Is(err, ErrAuthSecretRequired) {
		t.Fatalf("Parse() error = %v, want %v", err, ErrAuthSecretRequired)
	}
}

func TestResolveAuthSecret_UsesProvidedSecret(t *testing.T) {
	t.Setenv(EnvEnvironment, "")
	cfg := &Config{AuthSecret: "explicit"}
	if err := ResolveAuthSecret(cfg); err != nil {
		t.Fatalf("ResolveAuthSecret() error = %v", err)
	}
	if cfg.AuthSecret != "explicit" || cfg.EphemeralAuthSecret {
		t.Fatalf("cfg = %+v, want explicit secret", cfg)
	}
}

func TestResolveAuthSecret_NilConfig(t *testing.T) {
	if err := ResolveAuthSecret(nil); err == nil {
		t.Fatal("ResolveAuthSecret(nil) error = nil, want error")
	}
}

func TestResolveAuthSecret_EphemeralSecretsDiffer(t *testing.T) {
	t.Setenv(EnvEnvironment, "")
	first := &Config{}
	second := &Config{}
	if err := ResolveAuthSecret(first); err != nil {
		t.Fatalf("ResolveAuthSecret() error = %v", err)
	}
	if err := ResolveAuthSecret(second); err != nil {
		t.Fatalf("ResolveAuthSecret() error = %v", err)
	}
	if first.AuthSecret == "" || second.AuthSecret == "" {
		t.Fatal("ephemeral secret is empty")
	}
	if first.AuthSecret == second.AuthSecret {
		t.Fatal("ephemeral secrets must not be reused")
	}
	if !first.EphemeralAuthSecret || !second.EphemeralAuthSecret {
		t.Fatal("EphemeralAuthSecret = false, want true")
	}
}

func TestIsProduction(t *testing.T) {
	t.Setenv(EnvEnvironment, "production")
	if !IsProduction() {
		t.Fatal("IsProduction() = false, want true")
	}
	t.Setenv(EnvEnvironment, "prod")
	if !IsProduction() {
		t.Fatal("IsProduction(prod) = false, want true")
	}
	t.Setenv(EnvEnvironment, "dev")
	if IsProduction() {
		t.Fatal("IsProduction(dev) = true, want false")
	}
}

func TestEnvOverride(t *testing.T) {
	t.Setenv("TEST_CFG_ENV", "from-env")
	if got := envOverride("from-flag", "TEST_CFG_ENV"); got != "from-env" {
		t.Fatalf("envOverride() = %q, want from-env", got)
	}

	t.Setenv("TEST_CFG_ENV", "")
	if got := envOverride("from-flag", "TEST_CFG_ENV"); got != "from-flag" {
		t.Fatalf("envOverride() = %q, want from-flag", got)
	}
}
