package cassandradb

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/cassandra"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
)

// Migrator handles database schema migrations
// This encapsulates all migration logic and makes it testable
type Migrator struct {
	config config.DatabaseConfig
}

// NewMigrator creates a new migration manager
func NewMigrator(cfg config.DatabaseConfig) *Migrator {
	return &Migrator{config: cfg}
}

// Up runs all pending migrations
// This method ensures your database schema is always current
func (m *Migrator) Up() error {
	migrateInstance, err := m.createMigrate()
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer migrateInstance.Close()

	// Run migrations
	if err := migrateInstance.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// Down rolls back the last migration
// Useful for development and rollback scenarios
func (m *Migrator) Down() error {
	migrate, err := m.createMigrate()
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer migrate.Close()

	if err := migrate.Steps(-1); err != nil {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	return nil
}

// Version returns the current migration version
func (m *Migrator) Version() (uint, bool, error) {
	migrate, err := m.createMigrate()
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migrator: %w", err)
	}
	defer migrate.Close()

	return migrate.Version()
}

// createMigrate creates a migrate instance with proper configuration
func (m *Migrator) createMigrate() (*migrate.Migrate, error) {
	// Build Cassandra connection string
	// Format: cassandra://host:port/keyspace?consistency=level
	databaseURL := fmt.Sprintf("cassandra://%s/%s?consistency=%s",
		m.config.Hosts[0], // Use first host for migrations
		m.config.Keyspace,
		m.config.Consistency,
	)

	// Add authentication if configured
	if m.config.Username != "" && m.config.Password != "" {
		databaseURL = fmt.Sprintf("cassandra://%s:%s@%s/%s?consistency=%s",
			m.config.Username,
			m.config.Password,
			m.config.Hosts[0],
			m.config.Keyspace,
			m.config.Consistency,
		)
	}

	// Create migrate instance
	migrate, err := migrate.New(m.config.MigrationsPath, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize migrate: %w", err)
	}

	return migrate, nil
}
