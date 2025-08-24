package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/bensabler/milos-residence/internal/config"
	"github.com/bensabler/milos-residence/internal/driver"
	"github.com/bensabler/milos-residence/internal/forms"
	"github.com/bensabler/milos-residence/internal/helpers"
	"github.com/bensabler/milos-residence/internal/models"
	"github.com/bensabler/milos-residence/internal/render"
	"github.com/bensabler/milos-residence/internal/repository"
	"github.com/bensabler/milos-residence/internal/repository/dbrepo"
)

// Repo is the globally accessible handlers entrypoint used by the router.
var Repo *Repository

// Repository holds shared application dependencies for HTTP handlers.
type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

// NewRepo builds a Repository that shares the provided application config.
func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	// return a repository wired to the app config so handlers can access logs,
	// sessions, template cache, and other cross-cutting services
	return &Repository{
		App: a,
		DB:  dbrepo.NewPostgresRepo(db.SQL, a),
	}
}

// NewHandlers assigns the global Repo so package callers can reach handlers.
func NewHandlers(r *Repository) {
	// stash the repository so route wiring (e.g., handlers.Repo.Home) works
	Repo = r
}

// Home handles GET / by rendering the landing page.
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	m.DB.AllUsers()
	// render the home page with default template data (CSRF, flash, etc. added later)
	render.Template(w, r, "home.page.tmpl", &models.TemplateData{})
}

// About handles GET /about by rendering the about page.
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	// no dynamic data needed here yet—just render the static template
	render.Template(w, r, "about.page.tmpl", &models.TemplateData{})
}

// Photos handles GET /photos by rendering the gallery page.
func (m *Repository) Photos(w http.ResponseWriter, r *http.Request) {
	// serve the simple photo gallery page
	render.Template(w, r, "photos.page.tmpl", &models.TemplateData{})
}

// MakeReservation renders the reservation form page for GET /make-reservation.
func (m *Repository) MakeReservation(w http.ResponseWriter, r *http.Request) {
	// start with an empty reservation so the template has fields to bind to
	var emptyReservation models.Reservation
	data := make(map[string]interface{})
	data["reservation"] = emptyReservation

	// attach a fresh form wrapper so the template can show inline errors if present
	td := &models.TemplateData{
		Data:  data,
		Form:  forms.New(nil), // nil is fine here; template mainly reads Form.Errors
		Flash: "",
	}

	// render the reservation form for the visitor to complete
	render.Template(w, r, "make-reservation.page.tmpl", td)
}

// PostReservation processes a reservation form submission and redirects on success.
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	// parse the incoming form values so r.Form / r.PostForm are populated
	err := r.ParseForm()
	if err != nil {
		// if parsing fails, log it and return a 500 to the client
		helpers.ServerError(w, err)
		return
	}

	sd := r.Form.Get("start_date")
	ed := r.Form.Get("end_date")

	layout := "08-24-2025"
	startDate, err := time.Parse(layout, sd)
	if err != nil {
		helpers.ServerError(w, err)
	}
	endDate, err := time.Parse(layout, ed)
	if err != nil {
		helpers.ServerError(w, err)
	}

	roomID, err := strconv.Atoi(r.Form.Get("room_id"))
	if err != nil {
		helpers.ServerError(w, err)
	}

	// collect the posted fields into a typed Reservation for easier handling
	reservation := models.Reservation{
		FirstName: r.Form.Get("first_name"),
		LastName:  r.Form.Get("last_name"),
		Email:     r.Form.Get("email"),
		Phone:     r.Form.Get("phone"),
		StartDate: startDate,
		EndDate:   endDate,
		RoomID:    roomID,
	}

	// wrap the raw form values in our Form helper to run validations
	form := forms.New(r.PostForm)

	// ensure the core fields are present before proceeding
	form.Required("first_name", "last_name", "email", "phone")

	// enforce additional rules: first name length and email format
	form.MinLength("first_name", 3)
	form.IsEmail("email")

	// if any validation failed, re-render the form with inputs and inline errors
	if !form.Valid() {
		// prepare a data bag so the template can re-populate the user's entries
		data := make(map[string]interface{})
		data["reservation"] = reservation

		// show the form again with helpful messages beside each field
		render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	err = m.DB.InsertReservation(reservation)
	if err != nil {
		helpers.ServerError(w, err)
	}

	// on success, save the reservation in session for the summary page to read
	m.App.Session.Put(r.Context(), "reservation", reservation)

	// then send the user to the summary/confirmation page
	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

// GoldenHaybeamLoft shows the themed “golden haybeam” snooze spot page.
func (m *Repository) GoldenHaybeamLoft(w http.ResponseWriter, r *http.Request) {
	// render the dedicated feature page for this snooze spot
	render.Template(w, r, "golden-haybeam-loft.page.tmpl", &models.TemplateData{})
}

// WindowPerchTheater shows the themed “window perch theater” snooze spot page.
func (m *Repository) WindowPerchTheater(w http.ResponseWriter, r *http.Request) {
	// render the dedicated feature page for this snooze spot
	render.Template(w, r, "window-perch-theater.page.tmpl", &models.TemplateData{})
}

// LaundryBasketNook shows the themed “laundry basket nook” snooze spot page.
func (m *Repository) LaundryBasketNook(w http.ResponseWriter, r *http.Request) {
	// render the dedicated feature page for this snooze spot
	render.Template(w, r, "laundry-basket-nook.page.tmpl", &models.TemplateData{})
}

// Availability shows the availability search page (GET /search-availability).
func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	// render a simple form where users can input date ranges for availability
	render.Template(w, r, "search-availability.page.tmpl", &models.TemplateData{})
}

// PostAvailability handles the availability form submission and echoes the dates.
func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	// read the posted start and end dates from the form
	start := r.Form.Get("start")
	end := r.Form.Get("end")

	// respond with a simple, human-readable confirmation
	w.Write([]byte(fmt.Sprintf("start date is %s and end date is %s", start, end)))
}

// jsonResponse is the minimal payload we send back to AJAX callers.
type jsonResponse struct {
	OK      bool   `json:"ok"`      // true when the check passes (e.g., dates are available)
	Message string `json:"message"` // short, human-readable status for UI display
}

// AvailabilityJSON returns a small JSON payload indicating availability status.
func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	// define the payload we want to send back to the client
	resp := jsonResponse{
		OK:      true,
		Message: "Available!",
	}

	// marshal the response with indentation for readability
	out, err := json.MarshalIndent(resp, "", "     ")
	if err != nil {
		// if serialization fails, log the error and return a 500
		helpers.ServerError(w, err)
		return
	}

	// declare the content type and write the JSON body
	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

// Contact renders the contact page with ways to get in touch.
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	// show a simple page with contact details or a contact form
	render.Template(w, r, "contact.page.tmpl", &models.TemplateData{})
}

// ReservationSummary displays a confirmation page using the reservation from session.
func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	// attempt to retrieve the reservation stored during the POST step
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		// if it’s missing, log the issue, inform the user, and bounce to home
		m.App.ErrorLog.Println("Can't get error from session")
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// prevent stale data from lingering—clear the reservation from session
	m.App.Session.Remove(r.Context(), "reservation")

	// pass the reservation details to the template for display
	data := make(map[string]interface{})
	data["reservation"] = reservation

	// render the summary page with the collected data
	render.Template(w, r, "reservation-summary.page.tmpl", &models.TemplateData{
		Data: data,
	})
}
