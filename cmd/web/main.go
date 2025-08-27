package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/bensabler/milos-residence/internal/config"
	"github.com/bensabler/milos-residence/internal/driver"
	"github.com/bensabler/milos-residence/internal/handlers"
	"github.com/bensabler/milos-residence/internal/helpers"
	"github.com/bensabler/milos-residence/internal/models"
	"github.com/bensabler/milos-residence/internal/render"
)

// Application-wide dependencies using the Singleton pattern.
// These global variables provide a centralized point of access to shared resources
// across the entire application, ensuring consistent configuration and behavior.
var app config.AppConfig        // Central configuration store - implements Configuration pattern
var session *scs.SessionManager // Session manager - implements Session State pattern
var infoLog *log.Logger         // Operational logging - part of Observer pattern for system events
var errorLog *log.Logger        // Error logging with context - implements Error Handling pattern

// env implements the Environment Variable pattern with fallback defaults.
// This function encapsulates the logic for reading configuration values from the environment
// while providing safe defaults, preventing the application from panicking on missing config.
//
// Design Pattern: Strategy pattern - allows different configuration strategies (env vs defaults)
// Parameters:
//
//	key: the environment variable name to read
//	fallback: the default value if the environment variable is not set
//
// Returns: the environment value or the fallback if not found
func env(key, fallback string) string {
	// Attempt to read the environment variable
	if v := os.Getenv(key); v != "" {
		// Environment variable exists and has a value, return it
		return v
	}
	// Environment variable doesn't exist or is empty, return the safe default
	return fallback
}

// buildDSN implements the Builder pattern to construct a PostgreSQL Data Source Name.
// This function demonstrates how to safely construct complex configuration strings
// from multiple environment variables while handling optional components gracefully.
//
// Design Pattern: Builder pattern - constructs complex DSN string step by step
// Returns: a properly formatted PostgreSQL connection string
func buildDSN() string {
	// Collect required connection parameters with safe defaults
	// Each call to env() provides fallback values that work in development
	host := env("DB_HOST", "localhost") // Database server location
	port := env("DB_PORT", "5432")      // Standard PostgreSQL port
	user := env("DB_USER", "app")       // Database user account
	name := env("DB_NAME", "appdb")     // Target database name
	ssl := env("DB_SSLMODE", "disable") // SSL connection mode

	// Build the core connection parameters as key=value pairs
	// This approach makes the DSN construction transparent and maintainable
	parts := []string{
		"host=" + host,   // Server location
		"port=" + port,   // Connection port
		"user=" + user,   // Authentication user
		"dbname=" + name, // Target database
		"sslmode=" + ssl, // SSL configuration
	}

	// Handle optional password parameter
	// Only include password if it's actually set, avoiding empty password issues
	if pass := os.Getenv("DB_PASSWORD"); pass != "" {
		parts = append(parts, "password="+pass)
	}

	// Handle any additional connection parameters
	// This provides extensibility for future database configuration needs
	if extra := os.Getenv("DB_EXTRA"); extra != "" {
		parts = append(parts, extra)
	}

	// Join all parts into a single DSN string
	// PostgreSQL expects space-separated key=value pairs
	return strings.Join(parts, " ")
}

// main implements the Application Controller pattern and serves as the entry point.
// This function orchestrates the entire application startup sequence, demonstrating
// the Template Method pattern where the algorithm structure is defined but specific
// steps are delegated to other functions.
//
// Design Pattern: Template Method - defines the startup algorithm structure
// Design Pattern: Error Handling - ensures graceful failure and cleanup
func main() {
	// Delegate complex initialization to run() function following Single Responsibility Principle
	// This separation allows for better testing and error handling of the startup sequence
	db, err := run()
	if err != nil {
		// If initialization fails, terminate the application with a descriptive error
		// This implements the Fail Fast principle - better to crash early than corrupt data
		log.Fatal(err)
	}

	// Ensure database connection cleanup on application termination
	// This implements the Resource Management pattern using defer
	defer db.SQL.Close()

	// Configure HTTP server with application routes and middleware
	// The server address uses environment-based configuration with sensible defaults
	addr := ":" + env("PORT", "8080") // Default to port 8080 for development
	srv := &http.Server{
		Addr:    addr,         // Listen address and port
		Handler: routes(&app), // Route handler with middleware chain
	}

	// Log server startup information for operational monitoring
	// Include environment information to help distinguish between deployments
	infoLog.Printf("HTTP server listening on %s (env=%s)\n", addr, env("APP_ENV", "dev"))

	// Start the HTTP server and handle any startup errors
	// ListenAndServe blocks until the server stops or encounters an error
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		// Only treat unexpected errors as fatal - ErrServerClosed is normal during shutdown
		errorLog.Fatal(err)
	}
}

// run implements the Initialization pattern and Factory pattern for application setup.
// This function handles all complex startup logic including dependency injection,
// configuration loading, and service initialization. It demonstrates the
// Dependency Injection pattern by wiring together all application components.
//
// Design Pattern: Factory Method - creates and configures application dependencies
// Design Pattern: Dependency Injection - wires components together
// Returns: configured database connection and any initialization error
func run() (*driver.DB, error) {
	// Register data types for session storage using the Serialization pattern
	// The gob package needs to know about these types before they can be stored in sessions
	// This is required for the Session State pattern implementation
	gob.Register(models.Reservation{})     // Reservation data for booking flow
	gob.Register(models.User{})            // User data for authentication
	gob.Register(models.Room{})            // Room data for availability
	gob.Register(models.RoomRestriction{}) // Room restriction data for booking rules

	// Configure production vs development behavior using the Strategy pattern
	// This single flag controls multiple behaviors: cookie security, template caching, etc.
	app.InProduction = env("APP_ENV", "dev") == "prod"

	// Initialize logging infrastructure following the Observer pattern
	// These loggers will be used throughout the application to report events and errors

	// Create INFO logger for operational messages (startup, requests, etc.)
	// Standard format includes date and time for operational monitoring
	infoLog = log.New(os.Stdout, "INFO:\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog // Make available to other packages via dependency injection

	// Create ERROR logger with enhanced context (file and line numbers)
	// This helps developers quickly locate the source of errors during debugging
	errorLog = log.New(os.Stderr, "ERROR:\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog // Make available to other packages via dependency injection

	// Initialize session management using the Session State pattern
	// Sessions allow us to maintain user state across HTTP requests (which are stateless)
	session = scs.New()

	// Configure session behavior for security and usability
	session.Lifetime = 24 * time.Hour              // Session expires after 24 hours of inactivity
	session.Cookie.Persist = true                  // Session survives browser restart
	session.Cookie.SameSite = http.SameSiteLaxMode // CSRF protection while allowing navigation
	session.Cookie.Secure = app.InProduction       // HTTPS-only cookies in production for security

	// Make session manager available throughout the application
	app.Session = session

	// Establish database connection using the Connection Pool pattern
	// This creates a pool of reusable database connections for efficiency
	infoLog.Println("Connecting to database...")
	dsn := buildDSN() // Build connection string using Builder pattern
	db, err := driver.ConnectSQL(dsn)
	if err != nil {
		// Return wrapped error with context for better debugging
		return nil, fmt.Errorf("cannot connect to database: %s", err)
	}

	// Verify database connection and log connection details
	// This provides operational visibility into which database we're connected to
	var dbName, dbUser, schema, host string
	_ = db.SQL.QueryRow(`select current_database(), current_user, current_schema(), inet_server_addr()::text`).Scan(&dbName, &dbUser, &schema, &host)
	infoLog.Println("Connected to database")

	// Initialize template system using the Template Method pattern
	// Templates are parsed once and cached for performance (in production) or
	// rebuilt on each request (in development) for rapid iteration
	tc, err := render.CreateTemplateCache()
	if err != nil {
		// Template parsing failure is fatal - the application cannot serve pages
		return nil, fmt.Errorf("cannot create template cache: %s", err)
	}
	app.TemplateCache = tc // Store parsed templates in application config

	// Configure template caching behavior based on environment
	// This demonstrates the Strategy pattern - different caching strategies for different environments
	app.UseCache = env("USE_TEMPLATE_CACHE", "false") == "true"

	// Wire up application dependencies using Dependency Injection pattern
	// This creates the object graph where each component receives its dependencies

	// Create repository layer (Data Access Object pattern)
	repo := handlers.NewRepo(&app, db)

	// Initialize handlers with repository dependency
	handlers.NewHandlers(repo)

	// Configure template renderer with application config
	render.NewRenderer(&app)

	// Configure helper functions with application config
	helpers.NewHelpers(&app)

	// Return successfully configured database connection
	// The main function will use this for cleanup (defer db.SQL.Close())
	return db, nil
}
