// Command web is the application entry point. It bootstraps configuration,
// logging, sessions, database connectivity, template caching, handler wiring,
// and the HTTP server. Environment variables are used to configure runtime
// behavior (DB settings, ports, prod/dev flags).
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

// app holds the process-wide application configuration populated during startup.
var app config.AppConfig

// session is the global session manager used by middleware and handlers.
var session *scs.SessionManager

// infoLog writes informational messages; configured in run().
var infoLog *log.Logger

// errorLog writes error diagnostics; configured in run().
var errorLog *log.Logger

// env returns the environment variable value for key, or fallback if unset.
//
// Parameters:
//   - key: environment variable name.
//   - fallback: value returned when key is not set.
//
// Returns:
//   - string: resolved value.
//
// Usage:
//
//	dsn := env("DATABASE_URL", "postgres://user:pass@localhost:5432/app?sslmode=disable")
func env(key, fallback string) string {
	// Prefer explicit environment configuration; otherwise, default sensibly.
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// buildDSN constructs a PostgreSQL DSN string from individual environment
// variables. It supports an optional password and extra parameters.
//
// Reads:
//   - DB_HOST (default "localhost")
//   - DB_PORT (default "5432")
//   - DB_USER (default "app")
//   - DB_NAME (default "appdb")
//   - DB_SSLMODE (default "disable")
//   - DB_PASSWORD (optional)
//   - DB_EXTRA (optional; appended verbatim)
//
// Returns:
//   - string: DSN formatted for the pgx stdlib driver.
func buildDSN() string {
	// Resolve components from environment with safe defaults.
	host := env("DB_HOST", "localhost")
	port := env("DB_PORT", "5432")
	user := env("DB_USER", "app")
	name := env("DB_NAME", "appdb")
	ssl := env("DB_SSLMODE", "disable")

	// Assemble base parts; pgx accepts space-delimited key=value segments.
	parts := []string{
		"host=" + host,
		"port=" + port,
		"user=" + user,
		"dbname=" + name,
		"sslmode=" + ssl,
	}

	// Include optional password if provided.
	if pass := os.Getenv("DB_PASSWORD"); pass != "" {
		parts = append(parts, "password="+pass)
	}

	// Allow arbitrary extra params for advanced tuning (e.g., TimeZone=UTC).
	if extra := os.Getenv("DB_EXTRA"); extra != "" {
		parts = append(parts, extra)
	}

	return strings.Join(parts, " ")
}

// main coordinates process lifecycle: initialize subsystems, start the mail
// listener, build the HTTP server, and block on ListenAndServe. Fatal errors
// cause process exit.
//
// Side effects:
//   - Starts asynchronous mail listener.
//   - Logs server address and environment on startup.
//   - Defers database and mail channel cleanup.
func main() {
	// Perform full bootstrap and retrieve the live DB wrapper.
	db, err := run()
	if err != nil {
		log.Fatal(err)
	}
	defer db.SQL.Close() // ensure pool closes on shutdown

	// Close the mail channel after all senders are done.
	defer close(app.MailChan)

	// Start background email dispatcher (non-blocking).
	fmt.Println("Starting mail listener...")
	listenForMail()

	// Construct the HTTP server with resolved address and router.
	addr := ":" + env("PORT", "8080")
	srv := &http.Server{
		Addr:    addr,
		Handler: routes(&app),
	}

	// Announce server start with environment context.
	infoLog.Printf("HTTP server listening on %s (env=%s)\n", addr, env("APP_ENV", "dev"))

	// Serve until error or shutdown; ignore the normal ServerClosed signal.
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		errorLog.Fatal(err)
	}
}

// run performs application bootstrap and returns an initialized database handle.
// It configures gob, channels, environment flags, logging, sessions, database
// connectivity, template cache, repositories, renderers, and helpers.
//
// Returns:
//   - *driver.DB: initialized database wrapper.
//   - error: non-nil on unrecoverable bootstrap failure.
//
// Notes:
//   - Intended to make main() easy to test by isolating setup logic.
func run() (*driver.DB, error) {
	// Register complex types for session serialization.
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.RoomRestriction{})
	gob.Register(map[string]int{})

	// Initialize mail channel used by async sender.
	mailChan := make(chan models.MailData)
	app.MailChan = mailChan

	// Determine production mode from environment.
	app.InProduction = env("APP_ENV", "dev") == "prod"

	// Configure loggers with appropriate prefixes and flags.
	infoLog = log.New(os.Stdout, "INFO:\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog = log.New(os.Stderr, "ERROR:\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	// Configure secure cookie-backed session manager.
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction
	app.Session = session

	// Establish database connectivity.
	infoLog.Println("Connecting to database...")
	dsn := buildDSN()
	db, err := driver.ConnectSQL(dsn)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to database: %s", err)
	}
	infoLog.Println("Connected to database")

	// Build initial template cache.
	tc, err := render.CreateTemplateCache()
	if err != nil {
		return nil, fmt.Errorf("cannot create template cache: %s", err)
	}
	app.TemplateCache = tc

	// Toggle cache usage: typically true in production, false in development.
	app.UseCache = env("USE_TEMPLATE_CACHE", "false") == "true"

	// Wire repositories and package-level dependencies.
	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)
	render.NewRenderer(&app)
	helpers.NewHelpers(&app)

	return db, nil
}
