package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"strconv"
	"task-manager/pkg/config"
	"task-manager/pkg/logger"
)

func NewPostgres(cfg *config.Config) *sql.DB {
	port, err := strconv.Atoi(cfg.Postgres.Port)
	if err != nil {
		logger.Get().Fatal("Invalid PostgreSQL port",
			zap.String("port", cfg.Postgres.Port),
			zap.Error(err),
		)
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Postgres.Host,
		port,
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Name,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		logger.Get().Fatal("Failed to open Postgres connection",
			zap.Error(err),
		)
	}

	err = db.Ping()
	if err != nil {
		logger.Get().Fatal("Failed to ping Postgres",
			zap.Error(err),
			zap.String("dsn", fmt.Sprintf(
				"host=%s port=%d user=%s dbname=%s",
				cfg.Postgres.Host,
				port,
				cfg.Postgres.User,
				cfg.Postgres.Name,
			)),
		)
	}

	logger.Get().Info("Successfully connected to Postgres",
		zap.String("db", cfg.Postgres.Name),
	)

	return db
}
