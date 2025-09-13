// Package handlers provides HTTP handler setup and route wiring for the web application.
// This test setup file configures an isolated AppConfig, session manager, template cache,
// and router used by handler tests. It also supplies minimal utilities (mail listener,
// CSRF/session middleware, test template cache creator) required to exercise handlers.
package handlers

import (
	"encoding/gob"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/bensabler/milos-residence/internal/config"
	"github.com/bensabler/milos-residence/internal/helpers"
	"github.com/bensabler/milos-residence/internal/models"
	"github.com/bensabler/milos-residence/internal/render"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/justinas/nosurf"
)

// functions defines template helper functions available in test templates.
// These helpers are registered into the test template cache and mirror the
// production FuncMap so test templates execute identically.
//
// Keys:
//   - humanDate: formats a time as "01-02-2006"
//   - formatDate: formats a time using a supplied layout
//   - iterate: returns [0..count-1] for simple range loops
//   - add: returns a+b for index arithmetic inside templates
var functions = template.FuncMap{
	"humanDate":  func(t time.Time) string { return t.Format("01-02-2006") },
	"formatDate": func(t time.Time, f string) string { return t.Format(f) },
	"iterate": func(count int) []int {
		var items []int
		for i := 0; i < count; i++ {
			items = append(items, i)
		}
		return items
	},
	"add": func(a, b int) int { return a + b },
}

// app holds the application configuration scoped to tests.
// It is initialized in TestMain and injected into other packages (render, helpers)
// so handler code under test runs with deterministic settings.
var app config.AppConfig

// session manages per-request state during tests.
// Configured in TestMain with relaxed cookie/security settings appropriate for tests.
var session *scs.SessionManager

// pathToTemplates specifies the template root used by tests.
// Paths are relative to the test package directory.
var pathToTemplates = "./../../templates"

// TestMain bootstraps the test environment and runs all handler tests.
// It performs the following:
//   - Registers gob types used in session storage
//   - Configures logging, AppConfig, and scs session manager
//   - Starts a non-blocking mail listener on app.MailChan
//   - Builds and installs a template cache used during tests
//   - Initializes repositories, handlers, and helper/render packages
//   - Suppresses error log output for cleaner test output
func TestMain(m *testing.M) {
	// Register types for session encoding/decoding.
	gob.Register(models.Reservation{})

	// Configure application for test environment.
	app.InProduction = false

	// Set up logging.
	infoLog := log.New(os.Stdout, "INFO:\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog := log.New(os.Stdout, "ERROR:\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	// Configure session manager.
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction
	app.Session = session

	// Set up mail channel and start the listener to avoid blocking sends.
	mailChan := make(chan models.MailData)
	app.MailChan = mailChan
	defer close(mailChan)
	listenForMail()

	// Create and cache templates used by handler rendering paths.
	tc, err := CreateTestTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache: ", err)
	}
	app.TemplateCache = tc
	app.UseCache = true

	// Initialize repository and wire handlers/helpers/render to this AppConfig.
	repo := NewTestRepo(&app)
	NewHandlers(repo)
	render.NewRenderer(&app)
	helpers.NewHelpers(&app)

	// Suppress error log output during tests for cleaner output.
	errorLog.SetOutput(io.Discard)

	// Execute tests.
	os.Exit(m.Run())
}

// listenForMail drains app.MailChan to prevent test goroutines that send email
// from blocking. It runs for the lifetime of the test process.
func listenForMail() {
	go func() {
		for {
			_ = <-app.MailChan
		}
	}()
}

// getRoutes constructs the HTTP router configured for tests.
// It installs core middleware (panic recovery, CSRF, session) and registers
// all application routes against the test Repository.
//
// Returns:
//   - http.Handler: a chi.Mux with routes and middleware configured.
func getRoutes() http.Handler {
	mux := chi.NewRouter()

	// Core middleware used in tests to match production behavior.
	mux.Use(middleware.Recoverer)
	mux.Use(NoSurf)
	mux.Use(SessionLoad)

	// Public routes.
	mux.Get("/", Repo.Home)
	mux.Get("/about", Repo.About)
	mux.Get("/photos", Repo.Photos)

	mux.Get("/golden-haybeam-loft", Repo.GoldenHaybeamLoft)
	mux.Get("/window-perch-theater", Repo.WindowPerchTheater)
	mux.Get("/laundry-basket-nook", Repo.LaundryBasketNook)

	mux.Get("/search-availability", Repo.Availability)
	mux.Post("/search-availability", Repo.PostAvailability)
	mux.Post("/search-availability-json", Repo.AvailabilityJSON)

	mux.Get("/choose-room/{id}", Repo.ChooseRoom)
	mux.Get("/book-room", Repo.BookRoom)

	mux.Get("/contact", Repo.Contact)

	mux.Get("/make-reservation", Repo.MakeReservation)
	mux.Post("/make-reservation", Repo.PostReservation)
	mux.Get("/reservation-summary", Repo.ReservationSummary)

	// Auth.
	mux.Get("/user/login", Repo.ShowLogin)
	mux.Post("/user/login", Repo.PostShowLogin)
	mux.Get("/user/logout", Repo.Logout)

	// Static assets.
	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	// Admin routes (no auth middleware for tests).
	mux.Route("/admin", func(mux chi.Router) {
		mux.Get("/dashboard", Repo.AdminDashboard)
		mux.Get("/reservations-new", Repo.AdminNewReservations)
		mux.Get("/reservations-all", Repo.AdminAllReservations)
		mux.Get("/reservations-calendar", Repo.AdminReservationsCalendar)
		mux.Post("/reservations-calendar", Repo.AdminPostReservationsCalendar)
		mux.Get("/process-reservation/{src}/{id}/do", Repo.AdminProcessReservation)
		mux.Get("/delete-reservation/{src}/{id}/do", Repo.AdminDeleteReservation)
		mux.Get("/reservations/{src}/{id}/show", Repo.AdminShowReservation)
		mux.Post("/reservations/{src}/{id}", Repo.AdminPostShowReservation)
	})

	return mux
}

// NoSurf applies CSRF protection to the provided handler chain.
//
// Behavior:
//   - Uses nosurf with a base cookie set to HttpOnly, path "/", SameSite Lax,
//     and Secure honoring app.InProduction.
//
// Parameters:
//   - next: downstream handler to wrap with CSRF protection.
//
// Returns:
//   - http.Handler: the wrapped handler enforcing CSRF tokens.
func NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   app.InProduction,
		SameSite: http.SameSiteLaxMode,
	})

	return csrfHandler
}

// SessionLoad loads and saves the session for each request in the chain,
// ensuring handlers can read/write session state during tests.
//
// Parameters:
//   - next: downstream handler to wrap with session persistence.
//
// Returns:
//   - http.Handler: the wrapped handler with session lifecycle management.
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}

// CreateTestTemplateCache builds a template cache for tests by parsing all
// page (*.page.tmpl) and layout (*.layout.tmpl) templates rooted at pathToTemplates.
//
// Returns:
//   - map[string]*template.Template: compiled templates keyed by page filename
//   - error: non-nil on discovery or parse failure
func CreateTestTemplateCache() (map[string]*template.Template, error) {
	myCache := map[string]*template.Template{}

	// Find all page templates.
	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.tmpl", pathToTemplates))
	if err != nil {
		return myCache, err
	}

	// Parse each page template and attach any layouts.
	for _, page := range pages {
		name := filepath.Base(page)

		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return myCache, err
		}

		// Include layout templates if present.
		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))
		if err != nil {
			return myCache, err
		}

		if len(matches) > 0 {
			ts, err = ts.ParseGlob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))
			if err != nil {
				return myCache, err
			}
		}

		myCache[name] = ts
	}

	return myCache, nil
}
