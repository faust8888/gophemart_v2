package config

import (
	"errors"
	"flag"
	"testing"
)

func TestParse_Defaults(t *testing.T) {
	t.Setenv("RUN_ADDRESS", "")
	t.Setenv("DATABASE_URI", "")
	t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "")
	t.Setenv("AUTH_SECRET", "")

	cfg, err := Parse(nil)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if cfg.RunAddress != defaultRunAddress {
		t.Errorf("RunAddress = %q, want %q", cfg.RunAddress, defaultRunAddress)
	}
	if cfg.AuthSecret != defaultAuthSecret {
		t.Errorf("AuthSecret = %q, want %q", cfg.AuthSecret, defaultAuthSecret)
	}
	if cfg.DatabaseURI != "" {
		t.Errorf("DatabaseURI = %q, want empty", cfg.DatabaseURI)
	}
}

func TestParse_FlagsOverrideEnv(t *testing.T) {
	t.Setenv("RUN_ADDRESS", "env:1")
	t.Setenv("DATABASE_URI", "postgres://env")
	t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "http://env")

	cfg, err := Parse([]string{
		"-a", "flag:8081",
		"-d", "postgres://flag",
		"-r", "http://flag",
	})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if cfg.RunAddress != "flag:8081" {
		t.Errorf("RunAddress = %q, want flag value", cfg.RunAddress)
	}
	if cfg.DatabaseURI != "postgres://flag" {
		t.Errorf("DatabaseURI = %q, want flag value", cfg.DatabaseURI)
	}
	if cfg.AccrualAddress != "http://flag" {
		t.Errorf("AccrualAddress = %q, want flag value", cfg.AccrualAddress)
	}
}

func TestParse_EnvWhenFlagsEmpty(t *testing.T) {
	t.Setenv("RUN_ADDRESS", "localhost:9090")
	t.Setenv("DATABASE_URI", "postgres://user:pass@localhost/db")
	t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "http://accrual:8080")
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
