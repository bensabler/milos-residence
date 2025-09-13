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
// It initializes application configuration, session management, templates, and test repositories.
func TestMain(m *testing.M) {
	// Register types for session encoding/decoding
	gob.Register(models.Reservation{})

	// Configure application for test environment
	app.InProduction = false

	// Set up logging
	infoLog := log.New(os.Stdout, "INFO:\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog := log.New(os.Stdout, "ERROR:\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	// Configure session manager
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	// Set up mail channel
	mailChan := make(chan models.MailData)
	app.MailChan = mailChan
	defer close(mailChan)

	// Start mail listener
	listenForMail()

	// Create and cache templates
	tc, err := CreateTestTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache: ", err)
	}
	app.TemplateCache = tc
	app.UseCache = true

	// Initialize test repository and handlers
	repo := NewTestRepo(&app)
	NewHandlers(repo)
	render.NewRenderer(&app)
	helpers.NewHelpers(&app)

	// Suppress error log output during tests
	errorLog.SetOutput(io.Discard)

	// Execute tests
	os.Exit(m.Run())
}

// listenForMail consumes mail messages to prevent channel blocking during tests.
func listenForMail() {
	go func() {
		for {
			_ = <-app.MailChan
		}
	}()
}

// getRoutes creates and configures the HTTP router for testing.
// Returns a chi router with all application routes and middleware configured.
func getRoutes() http.Handler {
	mux := chi.NewRouter()

	// Apply middleware
	mux.Use(middleware.Recoverer)
	mux.Use(NoSurf)
	mux.Use(SessionLoad)

	// Configure routes
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

	mux.Get("/user/login", Repo.ShowLogin)
	mux.Post("/user/login", Repo.PostShowLogin)
	mux.Get("/user/logout", Repo.Logout)

	// Static files
	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	// Admin routes (no auth middleware for testing)
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

// SessionLoad wraps handlers with session loading and saving middleware.
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}

// CreateTestTemplateCache builds a template cache for testing.
// Parses all page and layout templates and returns a map of compiled templates.
func CreateTestTemplateCache() (map[string]*template.Template, error) {
	myCache := map[string]*template.Template{}

	// Find all page templates
	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.tmpl", pathToTemplates))
	if err != nil {
		return myCache, err
	}

	// Process each page template
	for _, page := range pages {
		name := filepath.Base(page)

		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return myCache, err
		}

		// Include layout templates if they exist
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
