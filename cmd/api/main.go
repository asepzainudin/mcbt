package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/asepzainudin14/mcbt/internal/config"
	"github.com/asepzainudin14/mcbt/internal/database"
	"github.com/asepzainudin14/mcbt/internal/pkg/logger"
	"github.com/asepzainudin14/mcbt/internal/server"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", slog.String("error", err.Error()))
		return err
	}

	log := logger.New(cfg.AppName, cfg.AppEnv, cfg.LogLevel)
	slog.SetDefault(log)

	if cfg.EnvFile != "" {
		log.Info("config loaded", slog.String("env_file", cfg.EnvFile))
	} else {
		log.Warn("no .env file found, using OS environment variables")
	}
	log.Info("database config",
		slog.String("host", cfg.DB.Host),
		slog.String("port", cfg.DB.Port),
		slog.String("name", cfg.DB.Name),
		slog.String("user", cfg.DB.User),
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.Connect(ctx, cfg, log)
	if err != nil {
		log.Error("failed to connect to database", slog.String("error", err.Error()))
		return err
	}

	router := server.NewRouter(server.RouterDeps{
		Cfg: cfg,
		Log: log,
		DB:  db,
	})

	srv := &http.Server{
		Addr:         cfg.AppHost + ":" + cfg.AppPort,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)

	go func() {
		log.Info("server starting", slog.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server failed", slog.String("error", err.Error()))
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		log.Info("shutdown signal received")
	case err := <-errCh:
		if err != nil {
			return err
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", slog.String("error", err.Error()))
		return err
	}

	if sqlDB, err := db.DB(); err == nil {
		_ = sqlDB.Close()
	}

	log.Info("server stopped")
	return nil
}
