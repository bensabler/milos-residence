package handlers

import (
	"encoding/gob"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/bensabler/milos-residence/internal/config"
	"github.com/bensabler/milos-residence/internal/models"
	"github.com/bensabler/milos-residence/internal/render"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/justinas/nosurf"
)

// functions holds template helper funcs that templates can call during tests.
var functions = template.FuncMap{}

// app is the test-scoped application configuration shared across helpers here.
var app config.AppConfig

// session is the session manager used by tests to mimic real request/session flows.
var session *scs.SessionManager

// pathToTemplates points to the templates directory relative to this test package.
var pathToTemplates = "./../../templates"

// getRoutes constructs a fully wired chi router for use by tests.
func getRoutes() http.Handler {
	// registers Reservation model data into the session
	gob.Register(models.Reservation{})

	// NOTE: Set to true for production to enable secure cookies and caching.
	app.InProduction = false

	// create and attach the info logger for general test diagnostics
	infoLog := log.New(os.Stdout, "INFO:\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	// create and attach the error logger with file/line info for easier debugging
	errorLog := log.New(os.Stdout, "ERROR:\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	// --- Session creation/configuration
	// create an in-memory session manager and configure cookie behavior
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	// store session in variable
	app.Session = session

	// --- Create template cache
	// build a fresh template cache from disk so tests render real templates
	tc, err := CreateTestTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache: ", err)
	}
	app.TemplateCache = tc

	// In production you typically set UseCache = true for performance.
	// tests prefer true here to exercise the cached code path deterministically
	app.UseCache = true

	// --- Dependency injection for handlers and render package
	// wire up the repository/handlers and renderer to the shared app config
	repo := NewRepo(&app)
	NewHandlers(repo)
	render.NewRenderer(&app)

	// create the chi router that tests will mount in an httptest server
	mux := chi.NewRouter()

	// Recover from panics with a 500 and stack trace in logs.
	mux.Use(middleware.Recoverer)
	// Add CSRF protection and session persistence.
	// mux.Use(NoSurf)
	mux.Use(SessionLoad)

	// Routes
	// core site pages
	mux.Get("/", Repo.Home)
	mux.Get("/about", Repo.About)
	mux.Get("/photos", Repo.Photos)

	// themed snooze-spot pages
	mux.Get("/golden-haybeam-loft", Repo.GoldenHaybeamLoft)
	mux.Get("/window-perch-theater", Repo.WindowPerchTheater)
	mux.Get("/laundry-basket-nook", Repo.LaundryBasketNook)

	// availability search (HTML + JSON)
	mux.Get("/search-availability", Repo.Availability)
	mux.Post("/search-availability", Repo.PostAvailability)
	mux.Post("/search-availability-json", Repo.AvailabilityJSON)

	// contact page
	mux.Get("/contact", Repo.Contact)

	// reservation flow
	mux.Get("/make-reservation", Repo.MakeReservation)
	mux.Post("/make-reservation", Repo.PostReservation)
	mux.Get("/reservation-summary", Repo.ReservationSummary)

	// serve static assets from ./static during tests, just like prod/dev
	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	// hand the configured router back to the caller (tests)
	return mux

}

// NoSurf wraps a handler with CSRF protection for all POST requests using nosurf.
// It sets a base cookie configured for HttpOnly and SameSite=Lax. In production,
// ensure app.InProduction is true so the cookie is marked Secure.
func NoSurf(next http.Handler) http.Handler {
	// wrap the downstream handler with nosurf's CSRF token validation
	csrfHandler := nosurf.New(next)

	// configure the CSRF cookie defaults used by nosurf
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   app.InProduction,
		SameSite: http.SameSiteLaxMode,
	})

	// return the protected handler back up the chain
	return csrfHandler
}

// SessionLoad loads and saves the session on every request.
// Place this middleware high in the chain so downstream handlers can access session data.
func SessionLoad(next http.Handler) http.Handler { return session.LoadAndSave(next) }

// CreateTemplateCache parses *.page.tmpl and *.layout.tmpl files into a cache keyed by page name.
func CreateTestTemplateCache() (map[string]*template.Template, error) {
	// start with an empty cache map for templates
	myCache := map[string]*template.Template{}

	// Collect all page templates.
	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.tmpl", pathToTemplates))
	if err != nil {
		return myCache, err
	}

	// Build a template set for each page and add layouts.
	for _, page := range pages {
		// use the base filename as the cache key
		name := filepath.Base(page)

		// create a new template, register helper funcs, and parse the page file
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return myCache, err
		}

		// look for any layout templates and attach them to the current template set
		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))
		if err != nil {
			return myCache, err
		}
		if len(matches) > 0 {
			if ts, err = ts.ParseGlob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates)); err != nil {
				return myCache, err
			}
		}

		// cache the assembled template set under its page name
		myCache[name] = ts
	}

	// hand the completed cache to the caller
	return myCache, nil
}
