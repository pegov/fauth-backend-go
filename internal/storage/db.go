package storage

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/pegov/fauth-backend-go/internal/log"
)

func GetDB(
	ctx context.Context,
	logger log.Logger,
	url string,
	maxIdleConns int,
	maxOpenConns int,
	connMaxLifetime time.Duration,
) (*sqlx.DB, error) {
	logger.Infof("Parsing DB config...")
	poolCfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}

	logger.Infof("Creating DB pool...")
	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, err
	}

	logger.Infof("Pinging DB...")
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}
	logger.Infof("DB is online!")

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
