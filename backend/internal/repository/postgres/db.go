// Package postgres implements every domain repository interface against
// PostgreSQL via GORM (CRUD paths) and raw pgx (bulk COPY / full-text search,
// where GORM would be either too slow or too awkward).
package postgres

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"bookreader/backend/internal/config"
)

func NewDB(cfg *config.Config) (*gorm.DB, error) {
	logLevel := logger.Warn
	if cfg.Env != "production" {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(cfg.Postgres.DSN), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logLevel),
		DisableForeignKeyConstraintWhenMigrating: true,
		PrepareStmt:                              true,
	})
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(cfg.Postgres.MaxOpenConn)
	sqlDB.SetMaxIdleConns(cfg.Postgres.MaxIdleConn)
	sqlDB.SetConnMaxLifetime(cfg.Postgres.ConnMaxLife)

	return db, nil
}
