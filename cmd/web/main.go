package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/bensabler/milos-residence/internal/config"
	"github.com/bensabler/milos-residence/internal/driver"
	"github.com/bensabler/milos-residence/internal/handlers"
	"github.com/bensabler/milos-residence/internal/helpers"
	"github.com/bensabler/milos-residence/internal/models"
	"github.com/bensabler/milos-residence/internal/render"
)

// app holds the application-wide configuration values and dependencies.
var app config.AppConfig

// session manages user sessions across requests via cookies.
var session *scs.SessionManager

// infoLog writes operational messages; errorLog records errors with context.
var infoLog *log.Logger
var errorLog *log.Logger

// env fetches an environment variable or returns the fallback.
// Avoids panics from missing configuration and centralizes defaults.
func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// buildDSN constructs a Postgres DSN from environment variables.
func buildDSN() string {
	host := env("DB_HOST", "localhost")
	port := env("DB_PORT", "5432")
	user := env("DB_USER", "app")
	pass := env("DB_PASS", "")
	name := env("DB_NAME", "appdb")
	ssl := env("DB_SSLMODE", "disable")
	extra := os.Getenv("DB_EXTRA") // optional, e.g. "search_path=public"

	base := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, pass, name, ssl)
	if extra != "" {
		return base + " " + extra
	}
	return base
}

// main is the entry point for the application. It initializes configuration
// (sessions, template cache, flags), registers handlers, and starts the HTTP server.
func main() {
	// call run to perform all setup work; bubble up any initialization error
	db, err := run()
	if err != nil {
		log.Fatal(err)
	}
	defer db.SQL.Close()

	addr := ":" + env("PORT", "8080")
	srv := &http.Server{
		Addr:    addr,
		Handler: routes(&app),
	}

	infoLog.Printf("HTTP server listening on %s (env=%s)\n", addr, env("APP_ENV", "dev"))
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		errorLog.Fatal(err)
	}
}

// run initializes configuration, sessions, template caching, and dependencies, and returns an error if setup fails.
func run() (*driver.DB, error) {
	// register types that will be stored in the session (so gob knows how to encode them)
	// what I am going to put in the session
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.RoomRestriction{})

	// set production flag; enable secure cookies and caching in production
	// NOTE: Set to true for production to enable secure cookies and caching.
	app.InProduction = env("APP_ENV", "dev") == "prod"

	// create an INFO logger for general operational messages
	infoLog = log.New(os.Stdout, "INFO:\t", log.Ldate|log.Ltime)
	// attach the info logger to app config so other packages can use it
	app.InfoLog = infoLog

	// create an ERROR logger with file/line info for debugging
	errorLog = log.New(os.Stderr, "ERROR:\t", log.Ldate|log.Ltime|log.Lshortfile)
	// attach the error logger to app config
	app.ErrorLog = errorLog

	// --- Session configuration
	// initialize the session manager
	session = scs.New()
	// set how long sessions last before expiring
	session.Lifetime = 24 * time.Hour
	// keep the session cookie across browser restarts
	session.Cookie.Persist = true
	// mitigate CSRF while allowing top-level navigation
	session.Cookie.SameSite = http.SameSiteLaxMode
	// only send cookies over HTTPS in production
	session.Cookie.Secure = app.InProduction
	// expose the session manager via app config
	app.Session = session

	// connect to database
	infoLog.Println("Connecting to database...")
	dsn := buildDSN()
	db, err := driver.ConnectSQL(dsn)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to database: %w", err)
	}
	infoLog.Println("Connected to database")

	// --- Template caching
	// build the template cache by parsing templates on disk
	tc, err := render.CreateTemplateCache()
	// if template parsing fails, log and abort startup
	if err != nil {
		return nil, fmt.Errorf("cannot create template cache: %w", err)
	}
	// store the parsed templates on the app config
	app.TemplateCache = tc

	// In production you typically set UseCache = true for performance.
	// choose whether to reuse the cache on each render (true in production)
	app.UseCache = env("USE_TEMPLATE_CACHE", "false") == "true"

	// --- Dependency injection for handlers and render package
	// wire handlers and support packages with the shared app config
	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)
	render.NewRenderer(&app)
	helpers.NewHelpers(&app)

	// all initialization succeeded
	return db, nil
}
