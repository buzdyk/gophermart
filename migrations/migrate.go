package migrations

import (
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(databaseDSN string, migrationsPath string) error {
	log.Println("Running database migrations...")

	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		databaseDSN,
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	srcErr, dbErr := m.Close()
	if srcErr != nil {
		return fmt.Errorf("error closing source: %w", srcErr)
	}

	if dbErr != nil {
		return fmt.Errorf("error closing database: %w", dbErr)
	}

	log.Println("Migrations applied successfully")

	return nil
}
