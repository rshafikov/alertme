package database

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBConnection struct {
	URL  string
	Pool *pgxpool.Pool
}

func NewDBConnection(dbURL string) *DBConnection {
	return &DBConnection{
		URL:  dbURL,
		Pool: nil,
	}
}

func (c *DBConnection) Connect(ctx context.Context) error {
	_, err := pgx.Connect(ctx, c.URL)
	if err != nil {
		return err
	}

	c.Pool, err = pgxpool.New(ctx, c.URL)
	if err != nil {
		return err
	}

	return nil
}

func (c *DBConnection) Ping(ctx context.Context) error {
	return c.Pool.Ping(ctx)
}
