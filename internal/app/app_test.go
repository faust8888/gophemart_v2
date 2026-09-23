package app

import (
	"errors"
	"testing"

	"github.com/faust8888/gophemart_v2/internal/config"
)

func TestRun_NilConfig(t *testing.T) {
	if err := Run(nil); err == nil {
		t.Fatal("Run(nil) error = nil, want error")
	}
}

func TestRun_EmptyDatabaseURI(t *testing.T) {
	err := Run(&config.Config{RunAddress: "localhost:0", AuthSecret: "secret"})
	if err == nil {
		t.Fatal("Run() error = nil, want error")
	}
}

func TestRun_ProductionRequiresAuthSecret(t *testing.T) {
	t.Setenv(config.EnvEnvironment, "production")
	err := Run(&config.Config{RunAddress: "localhost:0"})
	if !errors.Is(err, config.ErrAuthSecretRequired) {
		t.Fatalf("Run() error = %v, want %v", err, config.ErrAuthSecretRequired)
	}
}
