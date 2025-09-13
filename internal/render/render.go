// Package render centralizes HTML template rendering for the web layer.
// It manages a template cache (optionally disabled in development), injects
// default view data (CSRF token, flash messages, auth status), and exposes
// small template helpers via template.FuncMap. The package is configured at
// startup with an AppConfig and assumes templates live under pathToTemplates
// using *.page.tmpl and *.layout.tmpl naming conventions.
package render

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/bensabler/milos-residence/internal/config"
	"github.com/bensabler/milos-residence/internal/models"
	"github.com/justinas/nosurf"
)

// functions is the exported template helper map used by all parsed templates.
// Register new helpers here to make them available in *.tmpl files.
var functions = template.FuncMap{
	"humanDate":  HumanDate,
	"formatDate": FormatDate,
	"iterate":    Iterate,
	"add":        Add,
}

// app holds global application configuration and resources (logger, session,
// template cache, etc.). It is initialized once via NewRenderer during boot.
// Access to app is read-mostly at runtime; mutation should occur only at init.
var app *config.AppConfig

// pathToTemplates defines the on-disk location of template files. Override in
// tests or at startup when running from a different working directory.
var pathToTemplates = "./templates"

// Add returns the arithmetic sum of a and b.
// Typical usage is within templates that need index math.
func Add(a, b int) int { return a + b }

// Iterate returns a zero-based slice [0..count-1] to support simple loops in
// templates where a range over N items is needed.
func Iterate(count int) []int {
	var items []int
	for i := 0; i < count; i++ {
		items = append(items, i)
	}
	return items
}

// NewRenderer wires the render package to the provided AppConfig.
// It must be called during application initialization before any rendering.
func NewRenderer(a *config.AppConfig) {
	app = a
}

// HumanDate formats t as MM-DD-YYYY (01-02-2006), suitable for compact display
// in templates (e.g., lists, tables).
func HumanDate(t time.Time) string {
	return t.Format("01-02-2006")
}

// FormatDate returns t formatted with the provided layout f, which uses the
// Go time format reference "Mon Jan 2 15:04:05 MST 2006".
func FormatDate(t time.Time, f string) string {
	return t.Format(f)
}

// AddDefaultData injects standard cross-page data into td:
//   - Flash / Error / Warning: one-time messages popped from session
//   - CSRFToken: per-request token from nosurf
//   - IsAuthenticated: 1 if a user_id exists in session, otherwise 0
//
// Call this immediately before template execution to ensure dynamic values
// reflect the current request/session state.
func AddDefaultData(td *models.TemplateData, r *http.Request) *models.TemplateData {
	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Error = app.Session.PopString(r.Context(), "error")
	td.Warning = app.Session.PopString(r.Context(), "warning")
	td.CSRFToken = nosurf.Token(r)

	if app.Session.Exists(r.Context(), "user_id") {
		td.IsAuthenticated = 1
	}

	return td
}

// Template resolves and executes the named template into w using td as data.
// Behavior depends on configuration:
//   - If app.UseCache is true, it uses app.TemplateCache.
//   - Otherwise, it rebuilds a fresh cache by calling CreateTemplateCache.
//
// Errors are logged and mapped to generic HTTP 500 responses. A missing
// template key results in a concrete error ("can't get template from cache").
//
// Parameters:
//   - w: http.ResponseWriter to receive rendered output
//   - r: current request (used for session and CSRF)
//   - tmpl: template key (e.g., "home.page.tmpl")
//   - td: TemplateData to render (nil-safe; AddDefaultData will enrich it)
func Template(w http.ResponseWriter, r *http.Request, tmpl string, td *models.TemplateData) error {
	// Choose cache based on configuration.
	var (
		tc  map[string]*template.Template
		err error
	)

	if app.UseCache {
		tc = app.TemplateCache
	} else {
		tc, err = CreateTemplateCache()
		if err != nil {
			log.Printf("error creating template cache: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return err
		}
	}

	// Lookup the requested template.
	t, ok := tc[tmpl]
	if !ok {
		log.Printf("template %q not found in cache", tmpl)
		http.Error(w, "Template Not Found", http.StatusInternalServerError)
		return errors.New("can't get template from cache")
	}

	// Execute into a buffer to avoid partial writes on error.
	buf := new(bytes.Buffer)

	// Enrich request-specific defaults (flash, CSRF, auth flag, etc.).
	td = AddDefaultData(td, r)

	if err = t.Execute(buf, td); err != nil {
		log.Printf("error executing template %q: %v", tmpl, err)
		http.Error(w, "Template Execution Error", http.StatusInternalServerError)
		return err
	}

	// Write the full rendered payload.
	if _, err = buf.WriteTo(w); err != nil {
		fmt.Println("error writing template to response:", err)
	}
	return nil
}

// CreateTemplateCache parses all page and layout templates under pathToTemplates
// and returns a cache keyed by page template filename. Each entry is a compiled
// template with the shared helper FuncMap attached.
//
// Expected naming:
//   - Pages:  *.page.tmpl
//   - Layouts: *.layout.tmpl
//
// Returns a non-nil cache map on success. On failure, returns the partial map
// alongside the encountered error.
func CreateTemplateCache() (map[string]*template.Template, error) {
	myCache := map[string]*template.Template{}

	// Discover page templates.
	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.tmpl", pathToTemplates))
	if err != nil {
		return myCache, err
	}

	// Parse each page and its layouts into a single compiled template.
	for _, page := range pages {
		name := filepath.Base(page)

		// Start a new template with helpers.
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return myCache, err
		}

		// Attach any layout templates.
		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))
		if err != nil {
			return myCache, err
		}
		if len(matches) > 0 {
			if ts, err = ts.ParseGlob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates)); err != nil {
				return myCache, err
			}
		}

		myCache[name] = ts
	}

	return myCache, nil
}
