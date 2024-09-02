package storage

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func GetDB(
	ctx context.Context,
	logger *slog.Logger,
	url string,
	maxIdleConns int,
	maxOpenConns int,
	connMaxLifetime time.Duration,
) (*sqlx.DB, error) {
	logger.Info("Parsing DB config...")
	poolCfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}

	logger.Info("Creating DB pool...")
	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, err
	}

	logger.Info("Pinging DB...")
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}
	logger.Info("DB is online!")

	sqldb := stdlib.OpenDBFromPool(pool)
	sqldb.SetMaxIdleConns(maxIdleConns)
	sqldb.SetMaxOpenConns(maxOpenConns)
	sqldb.SetConnMaxLifetime(connMaxLifetime)

	db := sqlx.NewDb(sqldb, "pgx")

	sqlInit, err := os.ReadFile("./resources/sql/init.sql")
	if err != nil {
		return nil, fmt.Errorf("failed to read init.sql file: %w", err)
	}

	_, err = db.ExecContext(ctx, string(sqlInit))
	if err != nil {
		return nil, fmt.Errorf("failed to exec init.sql: %w", err)
	}

	return db, nil
}
