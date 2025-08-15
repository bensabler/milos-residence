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
	mux.Get("/photos", handlers.Repo.Photos)
	mux.Get("/room1", handlers.Repo.Room1)
	mux.Get("/room2", handlers.Repo.Room2)
	mux.Get("/room3", handlers.Repo.Room3)
	mux.Get("/search-availability", handlers.Repo.Availability)
	mux.Get("/contact", handlers.Repo.Contact)

	mux.Get("/reservation", handlers.Repo.Reservation)

	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}
