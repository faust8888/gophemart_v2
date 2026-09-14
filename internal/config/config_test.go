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
	t.Setenv("AUTH_SECRET", "")

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
	if cfg.AuthSecret != defaultAuthSecret {
		t.Errorf("AuthSecret = %q, want %q", cfg.AuthSecret, defaultAuthSecret)
	}
}

func TestParse_Flags(t *testing.T) {
	t.Setenv(EnvRunAddress, "")
	t.Setenv(EnvDatabaseURI, "")
	t.Setenv(EnvAccrualAddress, "")

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
	t.Setenv("AUTH_SECRET", "from-env")

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
