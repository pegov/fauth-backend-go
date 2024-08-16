package storage

import (
	"context"
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"

	"github.com/pegov/fauth-backend-go/internal/log"
)

func GetInMemoryDB(
	ctx context.Context,
	logger log.Logger,
	url string,
) (*sqlx.DB, error) {
	db, err := sqlx.Connect("sqlite3", url)
	if err != nil {
		return nil, err
	}

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
