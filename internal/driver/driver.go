// Package driver centralizes database connectivity and pooling configuration
// for the application. It exposes a thin wrapper (DB) around *sql.DB to enable
// dependency injection in other packages, and provides helpers to open and
// validate a PostgreSQL connection using the pgx driver.
//
// Usage:
//
//	conn, err := driver.ConnectSQL(os.Getenv("DATABASE_DSN"))
//	if err != nil { log.Fatal(err) }
//	defer conn.SQL.Close()
package driver

import (
	"database/sql"
	"time"

	// pgx stdlib driver and dependencies for database/sql.
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// DB wraps a *sql.DB so downstream code can depend on this type rather than
// the concrete driver. This allows repository packages to accept *driver.DB.
type DB struct {
	// SQL is the underlying connection pool.
	SQL *sql.DB
}

// dbConn holds the singleton DB instance returned by ConnectSQL.
// It is initialized once at application startup.
var dbConn = &DB{}

// Pool sizing and lifetime configuration. Tune these values based on workload
// and deployment characteristics (CPU, concurrency, DB server limits).
const (
	// maxOpenDbConn sets the maximum number of open connections to the database.
	maxOpenDbConn = 10
	// maxIdleDbConn sets the maximum number of idle connections in the pool.
	maxIdleDbConn = 5
	// maxDbLifetime sets the maximum amount of time a connection may be reused.
	maxDbLifetime = 5 * time.Minute
)

// ConnectSQL opens a PostgreSQL connection (via pgx), configures the pool,
// verifies connectivity, and returns the shared DB wrapper.
//
// Parameters:
//   - dsn: PostgreSQL DSN in the form accepted by pgx (e.g.,
//     "postgres://user:pass@host:5432/dbname?sslmode=disable").
//
// Returns:
//   - *DB: wrapper containing the configured *sql.DB pool
//   - error: non-nil if the connectivity test fails
//
// Notes:
//   - If NewDatabase fails, this function panics, matching the original
//     behavior so that startup fails fast on misconfiguration.
func ConnectSQL(dsn string) (*DB, error) {
	// Open the DB using the pgx driver via database/sql.
	d, err := NewDatabase(dsn)
	if err != nil {
		// Fail fast on early open errors (e.g., malformed DSN).
		panic(err)
	}

	// Apply pool tuning parameters.
	d.SetMaxOpenConns(maxOpenDbConn)
	d.SetMaxIdleConns(maxIdleDbConn)
	d.SetConnMaxLifetime(maxDbLifetime)

	// Publish the pool on the package-level wrapper.
	dbConn.SQL = d

	// Verify we can reach the database now (fail early in startup).
	if err = testDB(d); err != nil {
		return nil, err
	}

	return dbConn, nil
}

// testDB performs a simple Ping to validate database connectivity.
//
// Returns:
//   - error: non-nil if the database is unreachable or the connection is invalid.
func testDB(d *sql.DB) error {
	// Ping uses an existing or new connection to check liveness.
	if err := d.Ping(); err != nil {
		return err
	}
	return nil
}

// NewDatabase opens a new *sql.DB using the pgx stdlib driver and validates
// the connection with an initial Ping.
//
// Parameters:
//   - dsn: PostgreSQL DSN string.
//
// Returns:
//   - *sql.DB: an initialized connection pool
//   - error: non-nil if open or ping fails.
func NewDatabase(dsn string) (*sql.DB, error) {
	// sql.Open validates the driver name and prepares a lazy connector.
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	// Ensure the DSN is valid and the server is reachable.
	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
