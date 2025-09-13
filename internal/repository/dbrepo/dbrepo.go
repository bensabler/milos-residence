// Package dbrepo provides database repository implementations for the Milo's Residence application.
// It contains both production PostgreSQL implementations and test doubles for unit testing.
package dbrepo

import (
	"database/sql"

	"github.com/bensabler/milos-residence/internal/config"
	"github.com/bensabler/milos-residence/internal/repository"
)

// postgresDBRepo implements the DatabaseRepo interface using PostgreSQL.
// It holds database connection and application configuration for production operations.
type postgresDBRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}

// testDBRepo implements the DatabaseRepo interface for testing.
// It provides controlled responses without requiring a database connection.
type testDBRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}

// NewPostgresRepo creates a new PostgreSQL repository implementation.
// It requires an active database connection and application configuration.
// The returned repository is ready for production database operations.
func NewPostgresRepo(conn *sql.DB, a *config.AppConfig) repository.DatabaseRepo {
	return &postgresDBRepo{
		App: a,
		DB:  conn,
	}
}

// NewTestingRepo creates a new test repository implementation.
// It provides controlled responses for unit testing without requiring a database.
// The DB field is nil since no database connection is needed for testing.
func NewTestingRepo(a *config.AppConfig) repository.DatabaseRepo {
	return &testDBRepo{
		App: a,
	}
}
