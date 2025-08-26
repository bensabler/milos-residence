// Package main wires the HTTP router and middleware stack for the application.
package main

import (
	"net/http"

	"github.com/bensabler/milos-residence/internal/config"
	"github.com/bensabler/milos-residence/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// routes returns the top-level http.Handler for the application, including middleware.
func routes(app *config.AppConfig) http.Handler {
	// create a fresh router instance for this app
	mux := chi.NewRouter()

	// middleware stack — order matters; everything downstream flows through these
	// Recover from panics with a 500 and stack trace in logs.
	mux.Use(middleware.Recoverer)
	// Add CSRF protection and session persistence.
	mux.Use(NoSurf)
	mux.Use(SessionLoad)

	// Routes
	// core pages for the site
	mux.Get("/", handlers.Repo.Home)
	mux.Get("/about", handlers.Repo.About)
	mux.Get("/photos", handlers.Repo.Photos)

	// themed “snooze spot” detail pages
	mux.Get("/golden-haybeam-loft", handlers.Repo.GoldenHaybeamLoft)
	mux.Get("/window-perch-theater", handlers.Repo.WindowPerchTheater)
	mux.Get("/laundry-basket-nook", handlers.Repo.LaundryBasketNook)

	// availability search (HTML + JSON)
	mux.Get("/search-availability", handlers.Repo.Availability)
	mux.Post("/search-availability", handlers.Repo.PostAvailability)
	mux.Post("/search-availability-json", handlers.Repo.AvailabilityJSON)
	mux.Get("/choose-room/{id}", handlers.Repo.ChooseRoom)

	// contact page
	mux.Get("/contact", handlers.Repo.Contact)

	// reservation flow: show form → submit → show summary
	mux.Get("/make-reservation", handlers.Repo.MakeReservation)
	mux.Post("/make-reservation", handlers.Repo.PostReservation)
	mux.Get("/reservation-summary", handlers.Repo.ReservationSummary)

	// static assets served from ./static at /static/*
	fileServer := http.FileServer(http.Dir("./static/"))
	// map URLs beginning with /static to files under the local ./static directory
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	// return the configured router to be used by the HTTP server
	return mux
}
