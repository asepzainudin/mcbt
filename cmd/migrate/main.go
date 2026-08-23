package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/asepzainudin14/mcbt/internal/config"
	"github.com/asepzainudin14/mcbt/internal/pkg/logger"
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

	log := logger.New(cfg.AppName+"-migrator", cfg.AppEnv, cfg.LogLevel)

	command := flag.String("command", "up", "up | down | status")
	steps := flag.Int("steps", 1, "number of migrations to revert on down")
	flag.Parse()

	migrationsDir, err := findMigrationsDir()
	if err != nil {
		log.Error("failed to locate migrations directory", slog.String("error", err.Error()))
		return err
	}

	m, err := migrate.New("file://"+migrationsDir, migratorDSN(cfg))
	if err != nil {
		log.Error("failed to create migrator", slog.String("error", err.Error()))
		return err
	}
	defer func() {
		srcErr, dbErr := m.Close()
		for _, e := range []error{srcErr, dbErr} {
			if e != nil {
				log.Warn("migrator close warning", slog.String("error", e.Error()))
			}
		}
	}()

	switch *command {
	case "up":
		err = m.Up()
	case "down":
		err = m.Steps(-*steps)
	case "status":
		return printStatus(m, log)
	default:
		err = fmt.Errorf("unknown command %q (available: up, down, status)", *command)
	}

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Error("migration failed", slog.String("error", err.Error()))
		return err
	}

	version, dirty, _ := m.Version()
	if errors.Is(err, migrate.ErrNoChange) {
		log.Info("database already up to date")
	} else {
		log.Info("migration done",
			slog.Uint64("version", uint64(version)),
			slog.Bool("dirty", dirty),
		)
	}
	return nil
}

func printStatus(m *migrate.Migrate, log *slog.Logger) error {
	version, dirty, err := m.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			log.Info("no migrations applied yet")
			return nil
		}
		return err
	}
	log.Info("current migration version",
		slog.Uint64("version", uint64(version)),
		slog.Bool("dirty", dirty),
	)
	return nil
}

func findMigrationsDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	for {
		candidate := filepath.Join(dir, "migrations")
		if info, statErr := os.Stat(candidate); statErr == nil && info.IsDir() {
			return candidate, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New(
				"migrations directory not found: run from within the project or restore ./migrations",
			)
		}
		dir = parent
	}
}

func migratorDSN(cfg *config.Config) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.Name, cfg.DB.SSLMode,
	)
}
