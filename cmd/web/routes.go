// Command web wires HTTP routes and middleware for the application binary.
// It exposes the routes function, which constructs the chi router with
// global middleware, public endpoints, static assets, and admin subroutes.
package main

import (
	"net/http"

	"github.com/bensabler/milos-residence/internal/config"
	"github.com/bensabler/milos-residence/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// routes constructs the HTTP router and registers all endpoints.
//
// Behavior:
//   - Installs core middleware (panic recovery, CSRF protection, session load/save).
//   - Registers public site routes (home, about, rooms, availability, booking, auth).
//   - Serves static assets under /static/* from the local ./static directory.
//   - Nests admin routes under /admin protected by Auth middleware.
//
// Parameters:
//   - app: process-wide application configuration (unused here but kept for
//     symmetry and future expansion).
//
// Returns:
//   - http.Handler: a fully configured chi.Mux ready to pass to http.Server.
//
// Usage:
//
//	srv := &http.Server{Addr: addr, Handler: routes(app)}
func routes(app *config.AppConfig) http.Handler {
	mux := chi.NewRouter()

	// Core middleware — keep order logical: recover -> csrf -> session persistence.
	mux.Use(middleware.Recoverer)
	mux.Use(NoSurf)      // CSRF protection with nosurf base cookie policy in middleware.go
	mux.Use(SessionLoad) // scs session load/save wrapper

	// Public, non-auth routes.
	mux.Get("/", handlers.Repo.Home)
	mux.Get("/about", handlers.Repo.About)
	mux.Get("/photos", handlers.Repo.Photos)

	// Room detail pages.
	mux.Get("/golden-haybeam-loft", handlers.Repo.GoldenHaybeamLoft)
	mux.Get("/window-perch-theater", handlers.Repo.WindowPerchTheater)
	mux.Get("/laundry-basket-nook", handlers.Repo.LaundryBasketNook)

	// Availability search endpoints (HTML + JSON).
	mux.Get("/search-availability", handlers.Repo.Availability)
	mux.Post("/search-availability", handlers.Repo.PostAvailability)
	mux.Post("/search-availability-json", handlers.Repo.AvailabilityJSON)

	// Booking flow.
	mux.Get("/choose-room/{id}", handlers.Repo.ChooseRoom)
	mux.Get("/book-room", handlers.Repo.BookRoom)

	// Contact form.
	mux.Get("/contact", handlers.Repo.Contact)
	mux.Post("/contact", handlers.Repo.PostContact)

	// Reservation submission + confirmation.
	mux.Get("/make-reservation", handlers.Repo.MakeReservation)
	mux.Post("/make-reservation", handlers.Repo.PostReservation)
	mux.Get("/reservation-summary", handlers.Repo.ReservationSummary)

	// Authentication endpoints.
	mux.Get("/user/login", handlers.Repo.ShowLogin)
	mux.Post("/user/login", handlers.Repo.PostShowLogin)
	mux.Get("/user/logout", handlers.Repo.Logout)

	// Static assets served from local filesystem.
	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	// Admin routes — protected by Auth middleware, grouped under /admin.
	mux.Route("/admin", func(mux chi.Router) {
		mux.Use(Auth)

		mux.Get("/dashboard", handlers.Repo.AdminDashboard)

		mux.Get("/reservations-new", handlers.Repo.AdminNewReservations)
		mux.Get("/reservations-all", handlers.Repo.AdminAllReservations)
		mux.Get("/reservations-calendar", handlers.Repo.AdminReservationsCalendar)
		mux.Post("/reservations-calendar", handlers.Repo.AdminPostReservationsCalendar)
		mux.Get("/process-reservation/{src}/{id}/do", handlers.Repo.AdminProcessReservation)
		mux.Get("/delete-reservation/{src}/{id}/do", handlers.Repo.AdminDeleteReservation)

		mux.Get("/reservations/{src}/{id}/show", handlers.Repo.AdminShowReservation)
		mux.Post("/reservations/{src}/{id}", handlers.Repo.AdminPostShowReservation)
	})

	return mux
}
