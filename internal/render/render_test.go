// Package render provides tests validating default data injection, template
// rendering behavior, and template cache creation using the test AppConfig
// from setup_test.go.
package render

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bensabler/milos-residence/internal/models"
)

// TestAddDefaultData verifies that flash message values are popped from the
// session and placed onto TemplateData, ensuring one-time delivery semantics.
func TestAddDefaultData(t *testing.T) {
	var td models.TemplateData

	r, err := getSession()
	if err != nil {
		t.Error(err)
	}
	session.Put(r.Context(), "flash", "123")

	result := AddDefaultData(&td, r)

	if result.Flash != "123" {
		t.Error("flash value of 123 not found in session")
	}
}

// TestRenderTemplate exercises end-to-end template resolution and execution,
// including handling of a non-existent template key.
func TestRenderTemplate(t *testing.T) {
	// Use test-relative template path.
	pathToTemplates = "./../../templates"

	tc, err := CreateTemplateCache()
	if err != nil {
		t.Error(err)
	}

	app.TemplateCache = tc

	r, err := getSession()
	if err != nil {
		t.Error(err)
	}

	ww := httptest.NewRecorder()

	// Known existing template should render without error.
	if err = Template(ww, r, "home.page.tmpl", &models.TemplateData{}); err != nil {
		t.Error("error writing template to browser")
	}

	// Unknown template should yield an error.
	if err = Template(ww, r, "non-existent.page.tmpl", &models.TemplateData{}); err == nil {
		t.Error("rendered template that does not exist")
	}
}

// getSession creates a request bound to the test session context, enabling
// session reads/writes during handler and renderer tests.
func getSession() (*http.Request, error) {
	r, err := http.NewRequest("GET", "/some-url", nil)
	if err != nil {
		return nil, err
	}

	ctx := r.Context()
	ctx, _ = session.Load(ctx, r.Header.Get("X-Session"))

	return r.WithContext(ctx), nil
}

// TestNewTemplates confirms that NewRenderer assigns the package-global app
// configuration for subsequent render operations.
func TestNewTemplates(t *testing.T) {
	NewRenderer(app)
}

// TestCreateTemplateCache ensures page/layout discovery and parsing succeed
// against the test template directory.
func TestCreateTemplateCache(t *testing.T) {
	pathToTemplates = "./../../templates"

	if _, err := CreateTemplateCache(); err != nil {
		t.Error(err)
	}
}
