// Package main starts the HTTP server for a minimal web application.
// It wires application configuration, session management, templates, and routes.
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/bensabler/milos-residence/pkg/config"
	"github.com/bensabler/milos-residence/pkg/handlers"
	"github.com/bensabler/milos-residence/pkg/render"
)

// portNumber is the TCP address the HTTP server binds to.
// Use an environment variable for production deployments.
const portNumber = ":8080"

// app holds the application-wide configuration values and dependencies.
var app config.AppConfig

// session manages user sessions across requests via cookies.
var session *scs.SessionManager

// main is the entry point for the application. It initializes configuration
// (sessions, template cache, flags), registers handlers, and starts the HTTP server.
func main() {
	// NOTE: Set to true for production to enable secure cookies and caching.
	app.InProduction = false

	// --- Session configuration
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction
	app.Session = session

	// --- Template caching
	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache: ", err)
	}
	app.TemplateCache = tc

	// In production you typically set UseCache = true for performance.
	app.UseCache = false

	// --- Dependency injection for handlers and render package
	repo := handlers.NewRepo(&app)
	handlers.NewHandlers(repo)
	render.NewTemplates(&app)

	fmt.Printf("Starting application on port %s\n", portNumber)

	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}

	// Start the server; log.Fatal will exit on error (e.g., port in use).
	if err = srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
