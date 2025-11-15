package database

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rshafikov/alertme/internal/server/models"
	"testing"
)

func TestNewDB(t *testing.T) {
	var pool *pgxpool.Pool

	db := NewDB(pool)

	if db == nil {
		t.Error("Expected DB instance, got nil")
	}
	if db != nil && db.Pool != pool {
		t.Errorf("Expected pool %v, got %v", pool, db.Pool)
	}
}

func TestHandlePGErr(t *testing.T) {
	err := handlePGErr(nil, "test warning", "test code")
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
}

func TestDB_Add(t *testing.T) {
	t.Skip("Skipping TestDB_Add as it requires a real database connection")
	
	ctx := context.Background()
	db := &DB{}
	metric := &models.Metric{
		Name:  "test",
		Type:  models.GaugeType,
		Value: float64Ptr(1.0),
	}

	_ = db.Add(ctx, metric)
}

func TestDB_Get(t *testing.T) {
	t.Skip("Skipping TestDB_Get as it requires a real database connection")
	
	ctx := context.Background()
	db := &DB{}

	_, _ = db.Get(ctx, "", "")
}

func TestDB_List(t *testing.T) {
	t.Skip("Skipping TestDB_List as it requires a real database connection")
	
	ctx := context.Background()
	db := &DB{}
	_ = db.List(ctx)
}

func float64Ptr(f float64) *float64 {
	return &f
}

func TestDB_Clear(t *testing.T) {
	t.Skip("Skipping TestDB_Clear as it requires a real database connection")
	
	ctx := context.Background()
	db := &DB{}
	

	db.Clear(ctx)
}

func TestDB_AddBatch(t *testing.T) {
	t.Skip("Skipping TestDB_AddBatch as it requires a real database connection")
	
	ctx := context.Background()
	db := &DB{}
	metrics := []*models.Metric{
		{
			Name:  "test1",
			Type:  models.GaugeType,
			Value: float64Ptr(1.0),
		},
		{
			Name:  "test2",
			Type:  models.CounterType,
			Delta: func() *int64 { v := int64(10); return &v }(),
		},
	}
	
	_ = db.AddBatch(ctx, metrics)
}

func TestDB_Ping(t *testing.T) {
	t.Skip("Skipping TestDB_Ping as it requires a real database connection")
	
	ctx := context.Background()
	db := &DB{}
	
	_ = db.Ping(ctx)
}

func TestBootStrap(t *testing.T) {
	t.Skip("Skipping TestBootStrap as it requires a real database connection")
	
	ctx := context.Background()
	
	_, _ = BootStrap(ctx, "postgresql://localhost:5432/testdb")
}

func TestDBConnection_NewDBConnection(t *testing.T) {
	dbURL := "postgresql://localhost:5432/testdb"
	conn := NewDBConnection(dbURL)
	
	if conn.URL != dbURL {
		t.Errorf("Expected URL %s, got %s", dbURL, conn.URL)
	}
	
	if conn.Pool != nil {
		t.Error("Expected Pool to be nil initially")
	}
}

func TestMigrator_NewMigrator(t *testing.T) {
	var pool *pgxpool.Pool
	migrator := NewMigrator(pool)
	
	if migrator.Pool != pool {
		t.Errorf("Expected pool %v, got %v", pool, migrator.Pool)
	}
}

func TestMigrator_MakeMigrations(t *testing.T) {
	t.Skip("Skipping TestMigrator_MakeMigrations as it requires a real database connection")
	
	ctx := context.Background()
	var pool *pgxpool.Pool
	migrator := NewMigrator(pool)

	_ = migrator.MakeMigrations(ctx)
}