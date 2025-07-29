package database

import "context"

// Pinger is an interface for objects that can check database connectivity.
// Implementations should verify that the database connection is alive.
type Pinger interface {
	// Ping checks if the database connection is alive.
	// Returns an error if the connection is not available.
	Ping(ctx context.Context) error
}
