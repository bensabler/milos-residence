package render

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bensabler/milos-residence/internal/models"
)

// TestAddDefaultData verifies that AddDefaultData pulls values from session
// (e.g., flash messages) and merges them into TemplateData.
func TestAddDefaultData(t *testing.T) {
	// start with an empty template data bag
	var td models.TemplateData

	// build a request that carries a live session context
	r, err := getSession()
	if err != nil {
		t.Error(err)
	}

	// place a flash message into the session for this request
	session.Put(r.Context(), "flash", "123")

	// enrich td with defaults derived from the current request/session
	result := AddDefaultData(&td, r)

	// assert that the flash message was copied into the TemplateData
	if result.Flash != "123" {
		t.Error("flash value of 123 not found in session")
	}
}

// TestRenderTemplate ensures a known template renders without error,
// and that requesting a non-existent template returns an error.
func TestRenderTemplate(t *testing.T) {
	// point the renderer at the real templates directory used in tests
	pathToTemplates = "./../../templates"

	// build a template cache from disk
	tc, err := CreateTemplateCache()
	if err != nil {
		t.Error(err)
	}

	// install the cache into the test app configuration
	app.TemplateCache = tc

	// create a request with session context (needed for CSRF, messages, etc.)
	r, err := getSession()
	if err != nil {
		t.Error(err)
	}

	// use an in-memory response writer to capture output
	ww := httptest.NewRecorder()

	// render a known-good template; expect no error
	err = Template(ww, r, "home.page.tmpl", &models.TemplateData{})
	if err != nil {
		t.Error("error writing template to browser")
	}

	// attempt to render a template that doesn't exist; expect an error
	err = Template(ww, r, "non-existent.page.tmpl", &models.TemplateData{})
	if err == nil {
		t.Error("rendered template that does not exist")
	}
}

// getSession creates a request and attaches a session context so helpers
// like AddDefaultData can read/write session values during tests.
func getSession() (*http.Request, error) {
	// construct a basic GET request for testing
	r, err := http.NewRequest("GET", "/some-url", nil)
	if err != nil {
		return nil, err
	}

	// load or initialize the session and place it onto the request context
	ctx := r.Context()
	ctx, _ = session.Load(ctx, r.Header.Get("X-Session"))
	r = r.WithContext(ctx)

	// hand back the request carrying session state
	return r, nil
}

// TestNewTemplates confirms NewTemplates accepts the shared app config without error.
func TestNewTemplates(t *testing.T) {
	// call the initializer; it should not panic or fail
	NewRenderer(app)
}

// TestCreateTemplateCache verifies that templates on disk parse into a cache successfully.
func TestCreateTemplateCache(t *testing.T) {
	// point to the templates directory used by tests
	pathToTemplates = "./../../templates"

	// try to build the cache; any error indicates a parsing/globbing problem
	_, err := CreateTemplateCache()
	if err != nil {
		t.Error(err)
	}
}
