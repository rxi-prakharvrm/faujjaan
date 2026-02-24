package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 10
	cfg.MinConns = 1
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

func TableExists(ctx context.Context, pool *pgxpool.Pool, name string) (bool, error) {
	var regclass sql.NullString
	if err := pool.QueryRow(ctx, `SELECT to_regclass('public.' || $1)`, name).Scan(&regclass); err != nil {
		return false, err
	}
	return regclass.Valid && regclass.String != "", nil
}

