// Package database owns migration and seed helpers for the backend-api
// persistence baseline.
package database

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"time"

	migrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	migrationsDir = "migrations"
	seedsDir      = "seeds"
)

//go:embed migrations/*.sql seeds/*.sql
var schemaFiles embed.FS

// Migrator executes the repo-local migration and seed set against a Postgres
// database selected through DATABASE_URL.
type Migrator struct {
	databaseURL string
}

// NewMigrator constructs a migration runner bound to the target database URL.
func NewMigrator(databaseURL string) *Migrator {
	return &Migrator{databaseURL: databaseURL}
}

// Up applies all pending schema migrations.
func (m *Migrator) Up(ctx context.Context) error {
	runner, cleanup, err := m.runner(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	if err := runner.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}

// Down rolls back all applied schema migrations.
func (m *Migrator) Down(ctx context.Context) error {
	runner, cleanup, err := m.runner(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	if err := runner.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("rollback migrations: %w", err)
	}

	return nil
}

// Seed applies the deterministic baseline seed scripts after the schema exists.
func (m *Migrator) Seed(ctx context.Context) error {
	db, err := m.openDB(ctx)
	if err != nil {
		return err
	}
	defer db.Close()

	seedFiles, err := fs.Glob(schemaFiles, seedsDir+"/*.sql")
	if err != nil {
		return fmt.Errorf("list seed files: %w", err)
	}
	sort.Strings(seedFiles)

	for _, seedFile := range seedFiles {
		contents, err := schemaFiles.ReadFile(seedFile)
		if err != nil {
			return fmt.Errorf("read seed file %s: %w", seedFile, err)
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin seed transaction for %s: %w", seedFile, err)
		}

		if _, err := tx.ExecContext(ctx, string(contents)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("execute seed file %s: %w", seedFile, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit seed file %s: %w", seedFile, err)
		}
	}

	return nil
}

// Prepare applies all schema migrations and then runs the deterministic seed set.
func (m *Migrator) Prepare(ctx context.Context) error {
	if err := m.Up(ctx); err != nil {
		return err
	}

	if err := m.Seed(ctx); err != nil {
		return err
	}

	return nil
}

func (m *Migrator) runner(ctx context.Context) (*migrate.Migrate, func(), error) {
	db, err := m.openDB(ctx)
	if err != nil {
		return nil, nil, err
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		_ = db.Close()
		return nil, nil, fmt.Errorf("create postgres migration driver: %w", err)
	}

	source, err := iofs.New(schemaFiles, migrationsDir)
	if err != nil {
		_ = driver.Close()
		_ = db.Close()
		return nil, nil, fmt.Errorf("open embedded migrations: %w", err)
	}

	runner, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		_ = driver.Close()
		_ = db.Close()
		return nil, nil, fmt.Errorf("create migration runner: %w", err)
	}

	cleanup := func() {
		sourceErr, databaseErr := runner.Close()
		if sourceErr != nil || databaseErr != nil {
			_ = db.Close()
			return
		}
		_ = db.Close()
	}

	return runner, cleanup, nil
}

func (m *Migrator) openDB(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("pgx", m.databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open postgres database: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres database: %w", err)
	}

	return db, nil
}
