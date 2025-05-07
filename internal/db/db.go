package db

import (
	"database/sql"
	"log"

	_ "github.com/jackc/pgx/v4/stdlib" // PostgreSQL driver
)

func NewDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		log.Printf("Failed to ping database: %v", err)
		return nil, err
	}

	return db, nil
}
