package migrations

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed *.sql
var FS embed.FS

// RunMigrations applies all pending up migrations to the specified database URL.
func RunMigrations(databaseURL string) error {
	d, err := iofs.New(FS, ".")
	if err != nil {
		return fmt.Errorf("failed to create iofs migration driver: %w", err)
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database connection for migrations: %w", err)
	}
	defer db.Close()

	driver, err := pgx.WithInstance(db, &pgx.Config{})
	if err != nil {
		return fmt.Errorf("failed to create pgx migration driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", d, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to initialize migration instance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	log.Println("[migrations] Database schema is up to date")
	return nil
}

// RollbackMigration rolls back the last applied migration.
func RollbackMigration(databaseURL string) error {
	d, err := iofs.New(FS, ".")
	if err != nil {
		return fmt.Errorf("failed to create iofs migration driver: %w", err)
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database connection for migrations: %w", err)
	}
	defer db.Close()

	driver, err := pgx.WithInstance(db, &pgx.Config{})
	if err != nil {
		return fmt.Errorf("failed to create pgx migration driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", d, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to initialize migration instance: %w", err)
	}

	if err := m.Steps(-1); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	log.Println("[migrations] Rollback completed successfully")
	return nil
}
