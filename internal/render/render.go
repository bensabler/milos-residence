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

var functions = template.FuncMap{}

// app is the package-level reference to AppConfig set by NewTemplates.
var app *config.AppConfig
var pathToTemplates = "./templates"

// NewTemplates stores the application config for use by the render package.
func NewTemplates(a *config.AppConfig) { app = a }

// AddDefaultData injects default values into every template render (e.g., CSRF tokens,
// flash messages). To supply a CSRF token from nosurf, consider changing the signature to
// accept *http.Request and pulling nosurf.Token(r) here.
func AddDefaultData(td *models.TemplateData, r *http.Request) *models.TemplateData {
	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Error = app.Session.PopString(r.Context(), "error")
	td.Warning = app.Session.PopString(r.Context(), "warning")
	td.CSRFToken = nosurf.Token(r)
	return td
}

// RenderTemplate looks up and executes a named template, writing the result to w.
// When UseCache is false, it rebuilds the template cache on each request (handy in dev).
func RenderTemplate(w http.ResponseWriter, r *http.Request, tmpl string, td *models.TemplateData) error {
	var (
		tc  map[string]*template.Template
		err error
	)

	// Get the template cache from the app config
	if app.UseCache {
		tc = app.TemplateCache
	} else {
		// This is used for testing, so that we rebuild the cache on every request
		if tc, err = CreateTemplateCache(); err != nil {
			log.Printf("error creating template cache: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return err
		}
	}

	// Retrieve the requested template from the cache.
	t, ok := tc[tmpl]
	if !ok {
		log.Printf("template %q not found in cache", tmpl)
		http.Error(w, "Template Not Found", http.StatusInternalServerError)
		return errors.New("can't get template from cache")
	}

	buf := new(bytes.Buffer)
	td = AddDefaultData(td, r)

	if err = t.Execute(buf, td); err != nil {
		log.Printf("error executing template %q: %v", tmpl, err)
		http.Error(w, "Template Execution Error", http.StatusInternalServerError)
		return err
	}

	if _, err = buf.WriteTo(w); err != nil {
		fmt.Println("error writing template to response:", err)
	}

	return nil
}

// CreateTemplateCache parses *.page.tmpl and *.layout.tmpl files into a cache keyed by page name.
func CreateTemplateCache() (map[string]*template.Template, error) {
	myCache := map[string]*template.Template{}

	// Collect all page templates.
	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.tmpl", pathToTemplates))
	if err != nil {
		return myCache, err
	}

	// Build a template set for each page and add layouts.
	for _, page := range pages {
		name := filepath.Base(page)

		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return myCache, err
		}

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
