// Package app собирает зависимости и запускает HTTP-сервис накопительной системы лояльности.
package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/faust8888/gophemart_v2/internal/accrual"
	"github.com/faust8888/gophemart_v2/internal/auth"
	"github.com/faust8888/gophemart_v2/internal/config"
	"github.com/faust8888/gophemart_v2/internal/handler"
	"github.com/faust8888/gophemart_v2/internal/repository/postgres"
	"github.com/faust8888/gophemart_v2/internal/service"
	"github.com/faust8888/gophemart_v2/internal/worker"
)

const (
	tokenTTL          = 24 * time.Hour
	shutdownTimeout   = 10 * time.Second
	readHeaderTimeout = 5 * time.Second
)

// Run инициализирует хранилище, HTTP-сервер и обслуживает запросы до сигнала завершения.
func Run(cfg *config.Config) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	store, err := postgres.New(ctx, cfg.DatabaseURI)
	if err != nil {
		return err
	}
	defer store.Close()

	tokens, err := auth.NewManager(cfg.AuthSecret, tokenTTL)
	if err != nil {
		return err
	}

	users := service.NewUserService(store, tokens)
	orders := service.NewOrderService(store)
	balances := service.NewBalanceService(store)

	runCtx, cancelRun := context.WithCancel(ctx)
	defer cancelRun()

	pollerDone := make(chan struct{})
	if cfg.AccrualAddress != "" {
		poller := worker.NewAccrualPoller(
			store,
			accrual.NewClient(cfg.AccrualAddress, nil),
			worker.DefaultPollInterval,
		)
		go func() {
			defer close(pollerDone)
			poller.Run(runCtx)
		}()
	} else {
		close(pollerDone)
	}

	srv := &http.Server{
		Addr:              cfg.RunAddress,
		Handler:           handler.New(users, orders, balances, tokens).Routes(),
		ReadHeaderTimeout: readHeaderTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("gophermart started", slog.String("addr", cfg.RunAddress))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	waitPoller := func() {
		cancelRun()
		select {
		case <-pollerDone:
		case <-time.After(shutdownTimeout):
		}
	}

	select {
	case <-ctx.Done():
		waitPoller()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		waitPoller()
		return err
	}
}
