package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andrearcaina/draftkings-odds/internal/config"
	"github.com/andrearcaina/draftkings-odds/internal/draftkings"
	"github.com/andrearcaina/draftkings-odds/internal/logger"
	"github.com/andrearcaina/draftkings-odds/internal/web"
)

func main() {
	logger := logger.New()
	cfg, err := config.Load()
	if err != nil {
		logger.Error(context.Background(), "invalid configuration", "error", err)
		os.Exit(1)
	}

	client := draftkings.NewClient(cfg.DraftKingsURL, cfg.RefreshInterval, cfg.HTTPTimeout, logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go client.Run(ctx)

	server := &http.Server{
		Addr:              cfg.Addr,
		Handler:           web.NewHandler(client, logger, cfg.RefreshInterval).Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_ = server.Shutdown(shutdownCtx)
	}()

	logger.Info(context.Background(), "server listening", "addr", cfg.Addr, "refresh_interval_seconds", int64(cfg.RefreshInterval/time.Second))

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error(context.Background(), "server stopped", "error", err)
		os.Exit(1)
	}
}
