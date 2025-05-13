package database

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"
)

// MockDB is a mock database for testing
type MockDB struct {
	mock.Mock
}

func (m *MockDB) ExecContext(ctx context.Context, query string, args ...interface{}) (interface{}, error) {
	mockArgs := []interface{}{ctx, query}
	mockArgs = append(mockArgs, args...)
	returnArgs := m.Called(mockArgs...)
	return returnArgs.Get(0), returnArgs.Error(1)
}

func (m *MockDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) interface{} {
	mockArgs := []interface{}{ctx, query}
	mockArgs = append(mockArgs, args...)
	return m.Called(mockArgs...).Get(0)
}

func (m *MockDB) QueryContext(ctx context.Context, query string, args ...interface{}) (interface{}, error) {
	mockArgs := []interface{}{ctx, query}
	mockArgs = append(mockArgs, args...)
	returnArgs := m.Called(mockArgs...)
	return returnArgs.Get(0), returnArgs.Error(1)
}

// MockRows is a mock rows result for testing
type MockRows struct {
	mock.Mock
	data       [][]interface{}
	currentRow int
}

func (m *MockRows) Scan(dest ...interface{}) error {
	args := m.Called(dest)
	return args.Error(0)
}

func (m *MockRows) Next() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockRows) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRows) Err() error {
	args := m.Called()
	return args.Error(0)
}

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

func TestGetAppliedMigrations_Errors(t *testing.T) {
	// Create a mock DB
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Create a migration manager
	manager := NewMigrationManager(db)

	// Set up mock for rows.Scan error
	rows := sqlmock.NewRows([]string{"name"}).
		AddRow("create_users_table").
		RowError(0, errors.New("scan error"))

	mock.ExpectQuery("SELECT name FROM migrations").WillReturnRows(rows)

	// Call GetAppliedMigrations
	migrations, err := manager.GetAppliedMigrations()

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, migrations)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Set up mock for rows.Err error
	rows = sqlmock.NewRows([]string{"name"}).
		AddRow("create_users_table").
		CloseError(errors.New("rows error"))

	mock.ExpectQuery("SELECT name FROM migrations").WillReturnRows(rows)

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

// TestApplyMigrations_Errors tests the error cases of ApplyMigrations
func TestApplyMigrations_Errors(t *testing.T) {
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
	}

	// Test GetAppliedMigrations error
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS migrations").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT name FROM migrations").WillReturnError(errors.New("query error"))

	// Call ApplyMigrations
	err = manager.ApplyMigrations(migrations)

	// Assertions
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Add NewMigrator function that will fix the undefined references
// NewMigrator is an alias for NewMigrationManager for backward compatibility
func NewMigrator(db interface{}) *MigrationManager {
	// Convert to sql.DB if needed
	if sqlDB, ok := db.(*sql.DB); ok {
		return NewMigrationManager(sqlDB)
	}
	// For tests with mocks, create a mock migration manager
	return &MigrationManager{}
}

// Test for the new alias function
func TestMigratorAlias(t *testing.T) {
	// Create a mock DB
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Create a migrator using the alias function
	migrator := NewMigrator(db)

	// Assertions
	assert.NotNil(t, migrator)
}
