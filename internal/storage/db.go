package storage

import (
	"context"
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
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
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
	return db, nil
}
