package db

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func GetDB(dsn string) *sqlx.DB {
	log.Printf("Parsing DB config...")
	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		panic(err)
	}

	log.Printf("Creating DB pool...")
	pool, err := pgxpool.NewWithConfig(context.TODO(), poolCfg)
	if err != nil {
		panic(err)
	}

	log.Println("Pinging DB...")
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cancel()
	if err := pool.Ping(ctx); err != nil {
		panic(err)
	}
	log.Println("DB is online!")

	sqldb := stdlib.OpenDBFromPool(pool)
	sqldb.SetMaxIdleConns(4)
	sqldb.SetMaxOpenConns(8)
	sqldb.SetConnMaxLifetime(time.Minute * 10)

	db := sqlx.NewDb(sqldb, "pgx")
	return db
}
