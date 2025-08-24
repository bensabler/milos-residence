// Package render provides helpers for parsing and rendering HTML templates.
package render

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/bensabler/milos-residence/internal/config"
	"github.com/bensabler/milos-residence/internal/models"
	"github.com/justinas/nosurf"
)

// functions holds helper funcs that templates can call (e.g., formatters).
var functions = template.FuncMap{}

// app points to the shared application configuration set by NewTemplates.
var app *config.AppConfig

// pathToTemplates is where *.page.tmpl and *.layout.tmpl files live.
var pathToTemplates = "./templates"

// TODO(config): consider making this path configurable via env/flag for prod.

// NewRenderer stores the application config for use by the render package.
func NewRenderer(a *config.AppConfig) {
	// keep a package-level pointer to the shared app configuration
	app = a
}

// AddDefaultData injects standard values (flash, warning, error, CSRF) into TemplateData.
func AddDefaultData(td *models.TemplateData, r *http.Request) *models.TemplateData {
	// pull one-time messages from the session context (each read consumes the value)
	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Error = app.Session.PopString(r.Context(), "error")
	td.Warning = app.Session.PopString(r.Context(), "warning")

	// attach a CSRF token so forms rendered by this request can submit safely
	td.CSRFToken = nosurf.Token(r)

	// hand the enriched template data back to the caller
	return td
}

// RenderTemplate looks up and executes a named template, writing the result to w.
// When UseCache is false, it rebuilds the template cache per call (handy in dev/tests).
func Template(w http.ResponseWriter, r *http.Request, tmpl string, td *models.TemplateData) error {
	// set up local variables used during rendering
	var (
		// tc will hold the template cache we decide to use (prebuilt or freshly created)
		tc map[string]*template.Template

		// err captures any failure along the way (building the cache or executing the template)
		err error
	)

	// decide where to get templates from: cached in app, or rebuild on demand
	if app.UseCache {
		// fast path: reuse the prebuilt cache
		tc = app.TemplateCache
	} else {
		// dev/test path: rebuild the cache fresh to pick up template changes
		if tc, err = CreateTemplateCache(); err != nil {
			// log the problem, let the client know something went wrong, and stop
			log.Printf("error creating template cache: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return err
		}
	}

	// fetch the requested template from the cache by name
	t, ok := tc[tmpl]
	if !ok {
		// if it’s not there, log loudly and return a clear error
		log.Printf("template %q not found in cache", tmpl)
		http.Error(w, "Template Not Found", http.StatusInternalServerError)
		return errors.New("can't get template from cache")
	}

	// render into a buffer first so we can catch execution errors cleanly
	buf := new(bytes.Buffer)

	// add per-request defaults (CSRF, flash, etc.) before executing
	td = AddDefaultData(td, r)

	// execute the template with the provided data into the buffer
	if err = t.Execute(buf, td); err != nil {
		// log the exact template name and error for quick diagnosis
		log.Printf("error executing template %q: %v", tmpl, err)
		http.Error(w, "Template Execution Error", http.StatusInternalServerError)
		return err
	}

	// write the final HTML to the client
	if _, err = buf.WriteTo(w); err != nil {
		// writing failed after successful execution—log but don’t panic
		fmt.Println("error writing template to response:", err)
	}

	// everything went well
	return nil
}

// CreateTemplateCache parses *.page.tmpl and *.layout.tmpl files into a cache keyed by page name.
func CreateTemplateCache() (map[string]*template.Template, error) {
	// start with an empty cache map
	myCache := map[string]*template.Template{}

	// find all page templates on disk (each is a top-level view)
	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.tmpl", pathToTemplates))
	if err != nil {
		// if the glob itself failed, bubble the error up
		return myCache, err
	}

	// for every page template, build a template set and attach shared layouts
	for _, page := range pages {
		// use the base filename as the cache key (e.g., "home.page.tmpl")
		name := filepath.Base(page)

		// start a new template, register helper funcs, and parse the page file
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return myCache, err
		}

		// look for any matching layout templates and include them if present
		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))
		if err != nil {
			return myCache, err
		}
		if len(matches) > 0 {
			// ParseGlob merges any layouts into the current template set
			if ts, err = ts.ParseGlob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates)); err != nil {
				return myCache, err
			}
		}

		// store the assembled template set in the cache under the page name
		myCache[name] = ts
	}

	// return the completed cache for use by the renderer
	return myCache, nil
	// TODO(perf): in production, build this once at startup and set UseCache=true.
}
