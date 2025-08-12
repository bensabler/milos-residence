// Package handlers contains HTTP handler methods for the application.
// Handlers use a Repository which provides access to shared app dependencies.
package handlers

import (
	"net/http"

	"github.com/bensabler/milos-residence/pkg/config"
	"github.com/bensabler/milos-residence/pkg/models"
	"github.com/bensabler/milos-residence/pkg/render"
)

// Repo is the process-wide repository used by route handlers.
// NOTE: Prefer dependency injection over globals in larger applications.
var Repo *Repository

// Repository bundles application dependencies for handlers.
// Add more fields here as your app grows (e.g., DB connection).
type Repository struct {
	App *config.AppConfig
}

// NewRepo creates a new Repository bound to the provided AppConfig.
func NewRepo(a *config.AppConfig) *Repository {
	return &Repository{App: a}
}

// NewHandlers sets the package-level Repo used by the route layer.
func NewHandlers(r *Repository) { Repo = r }

// Home handles GET / by rendering the home page template.
// It stores the client's remote IP in session for demonstration purposes.
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	remoteIP := r.RemoteAddr
	m.App.Session.Put(r.Context(), "remote_ip", remoteIP)

	render.RenderTemplate(w, "home.page.tmpl", &models.TemplateData{})
}

// About handles GET /about by rendering the about page template.
// It demonstrates passing dynamic data into templates.
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)
	stringMap["test"] = "Hello, again."

	remoteIP := m.App.Session.GetString(r.Context(), "remote_ip")
	stringMap["remote_ip"] = remoteIP

	render.RenderTemplate(w, "about.page.tmpl", &models.TemplateData{StringMap: stringMap})
}
