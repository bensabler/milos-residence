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
	"github.com/bensabler/milos-residence/internal/handlers"
	"github.com/bensabler/milos-residence/internal/helpers"
	"github.com/bensabler/milos-residence/internal/models"
	"github.com/bensabler/milos-residence/internal/render"
)

// portNumber is the TCP address the HTTP server binds to.
// Use an environment variable for production deployments.
const portNumber = ":8080"

// app holds the application-wide configuration values and dependencies.
var app config.AppConfig

// session manages user sessions across requests via cookies.
var session *scs.SessionManager

// infoLog writes operational messages; errorLog records errors with context.
var infoLog *log.Logger
var errorLog *log.Logger

// main is the entry point for the application. It initializes configuration
// (sessions, template cache, flags), registers handlers, and starts the HTTP server.
func main() {
	// call run to perform all setup work; bubble up any initialization error
	err := run()
	if err != nil {
		log.Fatal(err)
	}

	// announce startup and the port we will bind to
	fmt.Printf("Starting application on port %s\n", portNumber)

	// construct the HTTP server with address and router handler
	srv := &http.Server{
		// listen on the configured port (e.g., ":8080")
		Addr: portNumber,
		// use the routes(...) function to provide the request handler
		Handler: routes(&app),
	}

	// start the server; exit if it returns an error (e.g., port already in use)
	if err = srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

// run initializes configuration, sessions, template caching, and dependencies, and returns an error if setup fails.
func run() error {
	// register types that will be stored in the session (so gob knows how to encode them)
	// what I am going to put in the session
	gob.Register(models.Reservation{})

	// set production flag; enable secure cookies and caching in production
	// NOTE: Set to true for production to enable secure cookies and caching.
	app.InProduction = false

	// create an INFO logger for general operational messages
	infoLog = log.New(os.Stdout, "INFO:\t", log.Ldate|log.Ltime)
	// attach the info logger to app config so other packages can use it
	app.InfoLog = infoLog

	// create an ERROR logger with file/line info for debugging
	errorLog = log.New(os.Stdout, "ERROR:\t", log.Ldate|log.Ltime|log.Lshortfile)
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

	// --- Template caching
	// build the template cache by parsing templates on disk
	tc, err := render.CreateTemplateCache()
	// if template parsing fails, log and abort startup
	if err != nil {
		log.Fatal("cannot create template cache: ", err)
		return err
	}
	// store the parsed templates on the app config
	app.TemplateCache = tc

	// In production you typically set UseCache = true for performance.
	// choose whether to reuse the cache on each render (true in production)
	app.UseCache = false

	// --- Dependency injection for handlers and render package
	// wire handlers and support packages with the shared app config
	repo := handlers.NewRepo(&app)
	handlers.NewHandlers(repo)
	render.NewTemplates(&app)
	helpers.NewHelpers(&app)

	// all initialization succeeded
	return nil
}
