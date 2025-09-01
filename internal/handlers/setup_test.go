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

// functions defines template helper functions for date formatting and iteration.
// These functions are available in all templates during testing.
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

// app holds the global application configuration for testing.
var app config.AppConfig

// session manages user session state during tests.
var session *scs.SessionManager

// pathToTemplates defines the relative path to HTML templates from the test directory.
var pathToTemplates = "./../../templates"

// TestMain sets up the test environment and runs all handler tests.
// It initializes the application configuration, session management, templates,
// and test repositories before executing the test suite.
func TestMain(m *testing.M) {
	// Register types for session encoding/decoding
	gob.Register(models.Reservation{})

	// Configure application for test environment (not production)
	app.InProduction = false

	// Set up logging for test output
	infoLog := log.New(os.Stdout, "INFO:\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog := log.New(os.Stdout, "ERROR:\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	// Configure session manager with test-appropriate settings
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	// Set up mail channel for testing (messages are consumed but not sent)
	mailChan := make(chan models.MailData)
	app.MailChan = mailChan
	defer close(mailChan)

	// Start mail listener to prevent channel blocking during tests
	listenForMail()

	// Create and cache templates for consistent test rendering
	tc, err := CreateTestTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache: ", err)
	}
	app.TemplateCache = tc

	// Enable template caching for consistent test behavior
	app.UseCache = true

	// Initialize test repository and handlers
	repo := NewTestRepo(&app)
	NewHandlers(repo)
	render.NewRenderer(&app)

	// Initialize helpers with app config to prevent nil pointer panics
	helpers.NewHelpers(&app)

	errorLog.SetOutput(io.Discard)

	// Execute all tests and exit with appropriate code
	os.Exit(m.Run())
}

// listenForMail consumes mail messages during tests to prevent channel blocking.
// This goroutine runs in the background, discarding all mail data sent to the channel.
func listenForMail() {
	go func() {
		for {
			// Consume mail messages without processing (testing mode)
			_ = <-app.MailChan
		}
	}()
}

// getRoutes creates and configures the HTTP router for testing.
// This function sets up all application routes with appropriate middleware
// and returns a configured chi router for use in test scenarios.
func getRoutes() http.Handler {
	// Create new chi router instance
	mux := chi.NewRouter()

	// Apply global middleware for request recovery, CSRF protection, and sessions
	mux.Use(middleware.Recoverer)
	mux.Use(NoSurf)
	mux.Use(SessionLoad)

	// Configure main informational page routes
	mux.Get("/", Repo.Home)
	mux.Get("/about", Repo.About)
	mux.Get("/photos", Repo.Photos)

	// Set up individual room detail page routes
	mux.Get("/golden-haybeam-loft", Repo.GoldenHaybeamLoft)
	mux.Get("/window-perch-theater", Repo.WindowPerchTheater)
	mux.Get("/laundry-basket-nook", Repo.LaundryBasketNook)

	// Configure availability search and booking workflow routes
	mux.Get("/search-availability", Repo.Availability)
	mux.Post("/search-availability", Repo.PostAvailability)
	mux.Post("/search-availability-json", Repo.AvailabilityJSON)

	mux.Get("/choose-room/{id}", Repo.ChooseRoom)
	mux.Get("/book-room", Repo.BookRoom)

	mux.Get("/contact", Repo.Contact)

	// Set up reservation creation and confirmation routes
	mux.Get("/make-reservation", Repo.MakeReservation)
	mux.Post("/make-reservation", Repo.PostReservation)
	mux.Get("/reservation-summary", Repo.ReservationSummary)

	// Configure user authentication routes
	mux.Get("/user/login", Repo.ShowLogin)
	mux.Post("/user/login", Repo.PostShowLogin)
	mux.Get("/user/logout", Repo.Logout)

	// Set up static file serving for CSS, JavaScript, and images
	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	// Configure administrative interface routes (auth middleware omitted for testing)
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

// NoSurf wraps handlers with CSRF protection middleware.
// This middleware generates and validates CSRF tokens for form submissions
// to prevent cross-site request forgery attacks.
func NoSurf(next http.Handler) http.Handler {
	// Create CSRF handler with the next middleware in chain
	csrfHandler := nosurf.New(next)

	// Configure CSRF cookie with security appropriate settings
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   app.InProduction,
		SameSite: http.SameSiteLaxMode,
	})

	return csrfHandler
}

// SessionLoad wraps handlers with session loading and saving middleware.
// This middleware automatically loads session data at request start
// and saves any changes when the request completes.
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}

// CreateTestTemplateCache builds a template cache for testing purposes.
// This function parses all page and layout templates, creating a map
// of compiled templates ready for use during test execution.
func CreateTestTemplateCache() (map[string]*template.Template, error) {
	// Initialize empty template cache
	myCache := map[string]*template.Template{}

	// Find all page template files using glob pattern
	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.tmpl", pathToTemplates))
	if err != nil {
		return myCache, err
	}

	// Process each page template individually
	for _, page := range pages {
		// Extract filename for cache key
		name := filepath.Base(page)

		// Parse page template with helper functions
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return myCache, err
		}

		// Find and include layout templates if they exist
		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))
		if err != nil {
			return myCache, err
		}

		// Parse layout templates if found
		if len(matches) > 0 {
			ts, err = ts.ParseGlob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))
			if err != nil {
				return myCache, err
			}
		}

		// Store compiled template in cache
		myCache[name] = ts
	}

	return myCache, nil
}
