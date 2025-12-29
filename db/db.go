package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"github.com/yookibooki/auth/config"
)

func Open(cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DB.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connection established")
	return db, nil
}

func Close(db *sql.DB) {
	if err := db.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
	}
}
