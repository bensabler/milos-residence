// Package handlers contains HTTP handler methods for the application.
// Handlers use a Repository which provides access to shared app dependencies.
package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/bensabler/milos-residence/internal/config"
	"github.com/bensabler/milos-residence/internal/forms"
	"github.com/bensabler/milos-residence/internal/models"
	"github.com/bensabler/milos-residence/internal/render"
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

	render.RenderTemplate(w, r, "home.page.tmpl", &models.TemplateData{})
}

// About handles GET /about by rendering the about page template.
// It demonstrates passing dynamic data into templates.
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)
	stringMap["test"] = "Hello, again."

	remoteIP := m.App.Session.GetString(r.Context(), "remote_ip")
	stringMap["remote_ip"] = remoteIP

	render.RenderTemplate(w, r, "about.page.tmpl", &models.TemplateData{StringMap: stringMap})
}

// Photos renders the photos page
func (m *Repository) Photos(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "photos.page.tmpl", &models.TemplateData{})
}

// Reservation renders the make a reservation page and displays form
func (m *Repository) MakeReservation(w http.ResponseWriter, r *http.Request) {
	var emptyReservation models.Reservation
	data := make(map[string]interface{})
	data["reservation"] = emptyReservation

	render.RenderTemplate(w, r, "make-reservation.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
		Data: data,
	})
}

// PostReservation handles the posting of a reservation form
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		return
	}

	reservation := models.Reservation{
		FirstName: r.Form.Get("first_name"),
		LastName: r.Form.Get("last_name"),
		Email: r.Form.Get("email"),
		Phone: r.Form.Get("phone"),
	}

	form := forms.New(r.PostForm)

	form.Required("first_name", "last_name", "email", "phone")
	form.MinLength("first_name", 3, r)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation

		render.RenderTemplate(w, r, "reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	m.App.Session.Put(r.Context(), "reservation", reservation)

	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

// GoldenHaybeamLoft renders the golden haybeam loft page
func (m *Repository) GoldenHaybeamLoft(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "golden-haybeam-loft.page.tmpl", &models.TemplateData{})
}

// WindowPerchTheater renders the window perch theater page
func (m *Repository) WindowPerchTheater(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "window-perch-theater.page.tmpl", &models.TemplateData{})
}

// LaundryBasketNook renders the laundry basket nook page
func (m *Repository) LaundryBasketNook(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "laundry-basket-nook.page.tmpl", &models.TemplateData{})
}

// Availability renders the search-availability page
func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "search-availability.page.tmpl", &models.TemplateData{})
}

// PostAvailability renders the search-availability page
func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	start := r.Form.Get("start")
	end := r.Form.Get("end")

	w.Write([]byte(fmt.Sprintf("start date is %s and end date is %s", start, end)))
}

type jsonResponse struct {
	OK bool `json:"ok"`
	Message string `json:"message"`
}

// AvailabilityJSON handles request for availability and send JSON response
func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	resp := jsonResponse {
		OK: true,
		Message: "Available!",
	}

	out, err := json.MarshalIndent(resp, "", "     ")
	if err != nil {
		log.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

// Contact renders the photos page
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "contact.page.tmpl", &models.TemplateData{})
}

func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		log.Println("cannot get item from session")
		return
	}

	data := make(map[string]interface{})
	data["reservation"] = reservation

	render.RenderTemplate(w, r, "reservation-summary.page.tmpl", &models.TemplateData{
		Data: data,
	})
}