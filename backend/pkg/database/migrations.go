package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// Migration represents a database migration
type Migration struct {
	ID        int64
	Name      string
	SQL       string
	Timestamp time.Time
}

// MigrationManager manages database migrations
type MigrationManager struct {
	db     *sql.DB
	logger *logrus.Entry
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(db *sql.DB) *MigrationManager {
	return &MigrationManager{
		db:     db,
		logger: logrus.WithField("component", "migration"),
	}
}

// Initialize creates the migrations table if it doesn't exist
func (m *MigrationManager) Initialize() error {
	// Create migrations table
	_, err := m.db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)

	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	m.logger.Info("Migration system initialized")
	return nil
}

// GetAppliedMigrations returns a list of applied migrations
func (m *MigrationManager) GetAppliedMigrations() (map[string]bool, error) {
	rows, err := m.db.Query("SELECT name FROM migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}
	defer rows.Close()

	appliedMigrations := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("failed to scan migration row: %w", err)
		}
		appliedMigrations[name] = true
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over migrations: %w", err)
	}

	return appliedMigrations, nil
}

// ApplyMigration applies a migration to the database
func (m *MigrationManager) ApplyMigration(migration Migration) error {
	// Begin transaction
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	// Execute migration SQL
	_, err = tx.Exec(migration.SQL)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to apply migration '%s': %w", migration.Name, err)
	}

	// Record migration
	_, err = tx.Exec("INSERT INTO migrations (name) VALUES ($1)", migration.Name)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to record migration '%s': %w", migration.Name, err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration '%s': %w", migration.Name, err)
	}

	m.logger.WithField("name", migration.Name).Info("Applied migration")
	return nil
}

// ApplyMigrations applies all migrations that haven't been applied yet
func (m *MigrationManager) ApplyMigrations(migrations []Migration) error {
	if err := m.Initialize(); err != nil {
		return err
	}

	appliedMigrations, err := m.GetAppliedMigrations()
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		if appliedMigrations[migration.Name] {
			m.logger.WithField("name", migration.Name).Debug("Migration already applied")
			continue
		}

		if err := m.ApplyMigration(migration); err != nil {
			return err
		}
	}

	m.logger.Info("All migrations applied")
	return nil
}
