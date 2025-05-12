package database

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMigrationManager(t *testing.T) {
	// Create a mock DB
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Create a migration manager
	manager := NewMigrationManager(db)

	// Assertions
	assert.NotNil(t, manager)
	assert.Equal(t, db, manager.db)
}

func TestInitialize(t *testing.T) {
	// Create a mock DB
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Create a migration manager
	manager := NewMigrationManager(db)

	// Set expectations for the migrations table creation
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS migrations").WillReturnResult(sqlmock.NewResult(0, 0))

	// Call Initialize
	err = manager.Initialize()

	// Assertions
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Test with error
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS migrations").WillReturnError(errors.New("db error"))

	// Call Initialize
	err = manager.Initialize()

	// Assertions
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAppliedMigrations(t *testing.T) {
	// Create a mock DB
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Create a migration manager
	manager := NewMigrationManager(db)

	// Set expectations for the query
	rows := sqlmock.NewRows([]string{"name"}).
		AddRow("create_users_table").
		AddRow("add_email_to_users")

	mock.ExpectQuery("SELECT name FROM migrations").WillReturnRows(rows)

	// Call GetAppliedMigrations
	migrations, err := manager.GetAppliedMigrations()

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, migrations, 2)
	assert.True(t, migrations["create_users_table"])
	assert.True(t, migrations["add_email_to_users"])
	assert.NoError(t, mock.ExpectationsWereMet())

	// Test with error
	mock.ExpectQuery("SELECT name FROM migrations").WillReturnError(errors.New("db error"))

	// Call GetAppliedMigrations
	migrations, err = manager.GetAppliedMigrations()

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, migrations)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestApplyMigration(t *testing.T) {
	// Create a mock DB
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Create a migration manager
	manager := NewMigrationManager(db)

	// Create a test migration
	migration := Migration{
		Name: "test_migration",
		SQL:  "CREATE TABLE test (id INT)",
	}

	// Set expectations for transaction
	mock.ExpectBegin()
	mock.ExpectExec("CREATE TABLE test").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO migrations").WithArgs("test_migration").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Call ApplyMigration
	err = manager.ApplyMigration(migration)

	// Assertions
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Test with error in migration
	mock.ExpectBegin()
	mock.ExpectExec("CREATE TABLE test").WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	// Call ApplyMigration
	err = manager.ApplyMigration(migration)

	// Assertions
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Test with error in insert
	mock.ExpectBegin()
	mock.ExpectExec("CREATE TABLE test").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO migrations").WithArgs("test_migration").WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	// Call ApplyMigration
	err = manager.ApplyMigration(migration)

	// Assertions
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestApplyMigrations(t *testing.T) {
	// Create a mock DB
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Create a migration manager
	manager := NewMigrationManager(db)

	// Create test migrations
	migrations := []Migration{
		{
			Name: "migration_1",
			SQL:  "CREATE TABLE test1 (id INT)",
		},
		{
			Name: "migration_2",
			SQL:  "CREATE TABLE test2 (id INT)",
		},
	}

	// Set expectations for initialize
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS migrations").WillReturnResult(sqlmock.NewResult(0, 0))

	// Set expectations for applied migrations query
	rows := sqlmock.NewRows([]string{"name"}).
		AddRow("migration_1")

	mock.ExpectQuery("SELECT name FROM migrations").WillReturnRows(rows)

	// Set expectations for the second migration only (since first is already applied)
	mock.ExpectBegin()
	mock.ExpectExec("CREATE TABLE test2").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO migrations").WithArgs("migration_2").WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectCommit()

	// Call ApplyMigrations
	err = manager.ApplyMigrations(migrations)

	// Assertions
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Test with error in initialize
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS migrations").WillReturnError(errors.New("db error"))

	// Call ApplyMigrations
	err = manager.ApplyMigrations(migrations)

	// Assertions
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Test with error in getting applied migrations
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS migrations").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT name FROM migrations").WillReturnError(errors.New("db error"))

	// Call ApplyMigrations
	err = manager.ApplyMigrations(migrations)

	// Assertions
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
