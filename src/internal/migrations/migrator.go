package migrations

import (
	"embed"
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed sql/*.sql
var sqlFiles embed.FS

// Up applies all migrations that have not been applied yet.
func Up() (runErr error) {
	connectionString := os.Getenv("DATABASE_CONNECTION_STRING")
	if connectionString == "" {
		return errors.New("DATABASE_CONNECTION_STRING is empty")
	}

	migrationURL, err := makeMigrationURL(connectionString)
	if err != nil {
		return err
	}

	sourceDriver, err := iofs.New(sqlFiles, "sql")
	if err != nil {
		return fmt.Errorf("create migration source: %w", err)
	}

	migrator, err := migrate.NewWithSourceInstance(
		"embedded-sql",
		sourceDriver,
		migrationURL,
	)
	if err != nil {
		_ = sourceDriver.Close()
		return fmt.Errorf("create migrator: %w", err)
	}

	defer func() {
		sourceErr, databaseErr := migrator.Close()
		runErr = errors.Join(runErr, sourceErr, databaseErr)
	}()

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}

func makeMigrationURL(connectionString string) (string, error) {
	databaseURL, err := url.Parse(connectionString)
	if err != nil {
		return "", fmt.Errorf("parse database URL: %w", err)
	}

	switch databaseURL.Scheme {
	case "postgres", "postgresql", "pgx5":
		databaseURL.Scheme = "pgx5"
	default:
		return "", fmt.Errorf("unsupported database scheme %q", databaseURL.Scheme)
	}

	return databaseURL.String(), nil
}
