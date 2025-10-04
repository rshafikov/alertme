// Package database provides functionality for connecting to and interacting with a PostgreSQL database.
package database

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBConnection represents a connection to a PostgreSQL database.
// It holds the connection URL and a connection pool for executing queries.
type DBConnection struct {
	// URL is the PostgreSQL connection string.
	URL string
	// Pool is the connection pool for executing queries.
	Pool *pgxpool.Pool
}

// NewDBConnection creates a new DBConnection with the specified database URL.
// The connection pool is initially nil and must be established by calling Connect.
func NewDBConnection(dbURL string) *DBConnection {
	return &DBConnection{
		URL:  dbURL,
		Pool: nil,
	}
}

// Connect establishes a connection to the database using the URL in the DBConnection.
// It first tests the connection and then creates a connection pool.
// Returns an error if the connection cannot be established.
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

// Ping checks if the database connection is alive.
// Returns an error if the connection is not available.
func (c *DBConnection) Ping(ctx context.Context) error {
	return c.Pool.Ping(ctx)
}
