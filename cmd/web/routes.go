// Package main (routes) wires up the HTTP router and middleware stack.
package main

import (
	"net/http"

	"github.com/bensabler/milos-residence/pkg/config"
	"github.com/bensabler/milos-residence/pkg/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// routes returns the top-level http.Handler for the application, including middleware.
func routes(app *config.AppConfig) http.Handler {
	mux := chi.NewRouter()

	// Recover from panics with a 500 and stack trace in logs.
	mux.Use(middleware.Recoverer)
	// Add CSRF protection and session persistence.
	mux.Use(NoSurf)
	mux.Use(SessionLoad)

	// Routes
	mux.Get("/", handlers.Repo.Home)
	mux.Get("/about", handlers.Repo.About)

	return mux
}
