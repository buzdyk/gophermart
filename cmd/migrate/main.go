package main

import (
	"flag"
	"log"
	"os"

	"github.com/riouske/gophermart/migrations"
)

func main() {
	var databaseDSN string
	var migrationsPath string

	flag.StringVar(&databaseDSN, "database", "", "Database connection string")
	flag.StringVar(&migrationsPath, "path", "./migrations", "Path to migrations directory")
	flag.Parse()

	if databaseDSN == "" {
		log.Fatal("Database connection string is required")
	}

	if err := migrations.RunMigrations(databaseDSN, migrationsPath); err != nil {
		log.Fatalf("Migration error: %v", err)
		os.Exit(1)
	}
}