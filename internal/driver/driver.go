package driver

import (
	"database/sql"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// DB implements the Connection Pool pattern for efficient database resource management.
// This struct wraps Go's standard database/sql connection pool, providing a clean
// interface for database operations while abstracting away the complexity of connection
// lifecycle management, pooling strategies, and resource cleanup. It demonstrates how
// production Go applications handle database connectivity with proper resource management.
//
// The Connection Pool pattern provides several critical benefits for web applications:
// 1. Resource efficiency - reuses database connections instead of creating new ones
// 2. Performance optimization - eliminates connection establishment overhead per request
// 3. Resource limiting - prevents database server overload through connection limits
// 4. Automatic cleanup - handles connection lifecycle and cleanup automatically
// 5. Fault tolerance - includes connection health monitoring and recovery
//
// Design Pattern: Connection Pool - manages database connection lifecycle and reuse
// Design Pattern: Resource Management - handles expensive resource allocation and cleanup
// Design Pattern: Proxy - provides simplified interface to complex database operations
type DB struct {
	// SQL holds the standard library connection pool for PostgreSQL operations
	// This *sql.DB instance manages multiple database connections internally,
	// providing thread-safe access to the database with automatic connection
	// pooling, health checking, and resource management handled by Go's runtime
	SQL *sql.DB
}

// dbConn implements the Singleton pattern for global database access.
// This package-level variable provides a single instance of the database connection
// pool that can be shared across the entire application. While global variables
// are generally discouraged, connection pools are a common exception because:
// 1. Database connections are expensive to create and should be reused
// 2. The connection pool is inherently thread-safe and designed for concurrent access
// 3. Most applications need exactly one connection pool per database
//
// The Singleton pattern here ensures that all parts of the application use the same
// connection pool, preventing resource waste and configuration inconsistencies.
var dbConn = &DB{}

// Connection pool configuration constants implement the Configuration pattern.
// These constants define the optimal database connection pool settings for production
// workloads, balancing resource usage with application performance requirements.
// The values are chosen based on common web application usage patterns and PostgreSQL
// server capabilities, but can be adjusted based on specific application needs.

// maxOpenDbConn limits the total number of open connections to the database.
// This prevents the application from overwhelming the database server with too many
// simultaneous connections, which could lead to resource exhaustion and performance
// degradation. The value of 10 is appropriate for small to medium web applications.
const maxOpenDbConn = 10

// maxIdleDbConn sets the number of idle connections kept in the pool.
// Idle connections can be immediately reused for new requests without the overhead
// of establishing a new connection. Setting this to 5 provides good performance
// while not consuming excessive database server resources during low-traffic periods.
const maxIdleDbConn = 5

// maxDbLifetime defines how long a connection can remain in the pool.
// After this duration, connections are automatically closed and recreated to prevent
// issues with stale connections, network timeouts, and database server maintenance.
// 5 minutes provides a good balance between connection reuse and freshness.
const maxDbLifetime = 5 * time.Minute

// ConnectSQL implements the Factory pattern for database connection creation.
// This function handles the complete process of establishing a database connection pool,
// configuring it with appropriate settings, and validating connectivity before returning
// it to the application. It demonstrates proper database initialization patterns including
// error handling, configuration, and health checking.
//
// The Factory pattern is appropriate here because database connection creation involves
// complex configuration, error handling, and validation that should be encapsulated
// in a single, reusable function rather than scattered throughout the application.
//
// Design Pattern: Factory Method - creates and configures database connection pools
// Design Pattern: Resource Management - handles connection pool lifecycle
// Design Pattern: Error Handling - comprehensive error checking and recovery
// Parameters:
//
//	dsn: Data Source Name containing database connection parameters
//
// Returns: Configured database instance ready for use, or error if connection fails
func ConnectSQL(dsn string) (*DB, error) {
	// Create new database connection using the Factory pattern
	// NewDatabase encapsulates the complexity of database driver selection,
	// connection string parsing, and initial connection establishment
	d, err := NewDatabase(dsn)
	if err != nil {
		// Database connection creation failed - this is typically a fatal error
		// during application startup because the application cannot function
		// without database connectivity. Using panic follows Go convention
		// for unrecoverable startup failures that should terminate the program
		panic(err)
	}

	// Configure connection pool parameters for production workloads
	// These settings implement best practices for web application database connectivity,
	// balancing performance requirements with resource conservation and server limits

	// Set maximum number of open connections to prevent database server overload
	// This limit ensures that the application cannot create more database connections
	// than the server can handle, preventing resource exhaustion under high load
	d.SetMaxOpenConns(maxOpenDbConn)

	// Configure idle connection pool for performance optimization
	// Maintaining idle connections eliminates connection establishment overhead
	// for subsequent requests, significantly improving response times
	d.SetMaxIdleConns(maxIdleDbConn)

	// Set connection lifetime to prevent stale connection issues
	// Automatic connection rotation prevents problems with network timeouts,
	// database server restarts, and other connectivity issues that can accumulate
	// over time in long-lived connections
	d.SetConnMaxLifetime(maxDbLifetime)

	// Store the configured connection pool in the global singleton instance
	// This makes the connection pool available throughout the application
	// while maintaining centralized configuration and lifecycle management
	dbConn.SQL = d

	// Validate database connectivity before returning the connection pool
	// This implements the Fail Fast principle - it's better to detect connectivity
	// problems during startup than to discover them during request processing
	err = testDB(d)
	if err != nil {
		// Connection validation failed - return error to caller for handling
		// This allows the application startup process to handle database
		// connectivity problems gracefully rather than panicking
		return nil, err
	}

	// Return successfully configured and validated database connection pool
	// The returned DB instance is ready for use by repository layers and
	// other components that need database access
	return dbConn, nil
}

// testDB implements the Health Check pattern for database connectivity validation.
// This function verifies that the database connection pool is functional by
// attempting a simple connectivity test. It demonstrates how production applications
// should validate external dependencies during startup to ensure system reliability.
//
// Health checks are critical for production systems because they enable:
// 1. Early detection of configuration problems during deployment
// 2. Automated monitoring and alerting for service health
// 3. Load balancer integration for traffic routing decisions
// 4. Graceful degradation when external dependencies are unavailable
//
// Design Pattern: Health Check - validates external dependency availability
// Design Pattern: Circuit Breaker - early detection of system failures
// Parameters:
//
//	d: Database connection pool to test
//
// Returns: nil if database is accessible, error if connectivity fails
func testDB(d *sql.DB) error {
	// Perform simple connectivity test using Ping method
	// Ping establishes a connection to the database server and verifies that
	// the connection is functional without performing any data operations.
	// This is the standard method for testing database connectivity in Go applications
	err := d.Ping()
	if err != nil {
		// Database connectivity test failed - return error with context
		// The error will include details about the specific connectivity problem,
		// such as network issues, authentication failures, or server unavailability
		return err
	}

	// Database connectivity test passed - connection pool is ready for use
	// Returning nil indicates successful validation and allows application
	// startup to continue with confidence in database availability
	return nil
}

// NewDatabase implements the Factory pattern for database driver initialization.
// This function handles the low-level details of database driver registration,
// connection string processing, and initial connection establishment. It abstracts
// away the PostgreSQL-specific implementation details behind a clean interface
// that could potentially support multiple database types in the future.
//
// The Factory pattern is particularly valuable for database connections because:
// 1. Driver registration and configuration vary significantly between database types
// 2. Connection string formats and requirements differ across database systems
// 3. Error handling and validation requirements are database-specific
// 4. Future database migration is simplified through abstraction
//
// Design Pattern: Factory Method - creates database connections with proper configuration
// Design Pattern: Abstract Factory - could potentially support multiple database types
// Parameters:
//
//	dsn: Data Source Name containing PostgreSQL connection parameters
//
// Returns: Configured sql.DB connection pool or error if creation fails
func NewDatabase(dsn string) (*sql.DB, error) {
	// Create database connection using PostgreSQL driver
	// The "pgx" driver is imported with blank identifier to register itself
	// with the sql package. This registration pattern is standard in Go
	// database applications and allows the sql.Open call to use the driver
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		// Database connection creation failed - return error to caller
		// This could be due to invalid DSN format, driver registration problems,
		// or other configuration issues that prevent connection pool creation
		return nil, err
	}

	// Validate that the database connection is actually functional
	// sql.Open only creates the connection pool structure - it doesn't actually
	// connect to the database until the first operation. Ping forces a real
	// connection attempt to validate that the DSN and database are accessible
	if err = db.Ping(); err != nil {
		// Database connectivity validation failed - return error with context
		// This indicates problems with network connectivity, authentication,
		// database availability, or other runtime connectivity issues
		return nil, err
	}

	// Return successfully created and validated database connection pool
	// The connection pool is now ready for use by the application with
	// proper driver registration, connection validation, and error handling
	return db, nil
}
