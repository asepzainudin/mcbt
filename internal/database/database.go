package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/asepzainudin14/mcbt/internal/config"
)

func Connect(ctx context.Context, cfg *config.Config, log *slog.Logger) (*gorm.DB, error) {
	logLevel := gormlogger.Warn
	if cfg.AppEnv == "production" {
		logLevel = gormlogger.Error
	}

	db, err := gorm.Open(postgres.Open(cfg.DB.DSN()), &gorm.Config{
		Logger:                 gormlogger.Default.LogMode(logLevel),
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.DB.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.DB.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime())

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(pingCtx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	log.Info("database connected",
		slog.String("host", cfg.DB.Host),
		slog.String("port", cfg.DB.Port),
		slog.String("name", cfg.DB.Name),
	)

	return db, nil
}
