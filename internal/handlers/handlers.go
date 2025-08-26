package handlers

import (
	"encoding/json"
	"errors"
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
	"github.com/go-chi/chi/v5"
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
	// Get the reservation from the session
	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		// Key not found or wrong type
		helpers.ServerError(w, errors.New("cannot get reservation from session"))
		return
	}

	room, err := m.DB.GetRoomByID(res.RoomID)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	res.Room.RoomName = room.RoomName

	m.App.Session.Put(r.Context(), "reservation", res)

	// format start and end date
	sd := res.StartDate.Format("01/02/2006")
	ed := res.EndDate.Format("01/02/2006")

	// insert start and end date data into map of strings
	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	data := make(map[string]interface{})
	data["reservation"] = res

	// attach a fresh form wrapper so the template can show inline errors if present
	td := &models.TemplateData{
		Data:      data,
		Form:      forms.New(nil), // nil is fine here; template mainly reads Form.Errors
		StringMap: stringMap,
	}

	// render the reservation form for the visitor to complete
	render.Template(w, r, "make-reservation.page.tmpl", td)
}

// PostReservation processes a reservation form submission and redirects on success.
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		helpers.ServerError(w, errors.New("cn't get from session"))
		return
	}

	// parse the incoming form values so r.Form / r.PostForm are populated
	err := r.ParseForm()
	if err != nil {
		// if parsing fails, log it and return a 500 to the client
		helpers.ServerError(w, err)
		return
	}

	reservation.FirstName = r.Form.Get("first_name")
	reservation.LastName = r.Form.Get("last_name")
	reservation.Phone = r.Form.Get("phone")
	reservation.Email = r.Form.Get("email")

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

	newReservationID, err := m.DB.InsertReservation(reservation)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	restriction := models.RoomRestriction{
		StartDate:     reservation.StartDate,
		EndDate:       reservation.EndDate,
		RoomID:        reservation.RoomID,
		ReservationID: newReservationID,
		RestrictionID: 1,
	}

	err = m.DB.InsertRoomRestriction(restriction)
	if err != nil {
		helpers.ServerError(w, err)
		return
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

	const layout = "01/02/2006"
	startDate, err := time.Parse(layout, start)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	endDate, err := time.Parse(layout, end)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// query db for all available rooms and store output in a slice of type model-Room
	rooms, err := m.DB.SearchAvailabilityForAllRooms(startDate, endDate)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// if slice of rooms is empty
	if len(rooms) == 0 {
		// no availability
		// add error to context and redirect to /search-availability
		m.App.Session.Put(r.Context(), "error", "No availability")
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
		return
	}

	// store the rooms in a data variable that is a map of string interfaces
	// assign key rooms to variable that holds the slice of available rooms
	data := make(map[string]interface{})
	data["rooms"] = rooms

	// read the start and end date values from the form data into a Reservation model and stored in res
	res := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
	}

	// putting the start and end date data from the posted form in the session
	m.App.Session.Put(r.Context(), "reservation", res)

	// render the choose-room template and pass the room data
	render.Template(w, r, "choose-room.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

// jsonResponse is the minimal payload we send back to AJAX callers.
type jsonResponse struct {
	OK        bool   `json:"ok"`      // true when the check passes (e.g., dates are available)
	Message   string `json:"message"` // short, human-readable status for UI display
	RoomID    string `json:"room_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// AvailabilityJSON returns a small JSON payload indicating availability status.
func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	sd := r.Form.Get("start")
	ed := r.Form.Get("end")

	layout := "01/02/2006"
	startDate, _ := time.Parse(layout, sd)
	endDate, _ := time.Parse(layout, ed)

	roomID, _ := strconv.Atoi(r.Form.Get("room_id"))

	available, err := m.DB.SearchAvailabilityByDatesByRoomID(startDate, endDate, roomID)
	resp := jsonResponse{
		OK:        available,
		Message:   "",
		StartDate: sd,
		EndDate:   ed,
		RoomID:    strconv.Itoa(roomID),
	}
	if err != nil {
		helpers.ServerError(w, err)
		return
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

	sd := reservation.StartDate.Format("01/02/2006")
	ed := reservation.EndDate.Format("01/02/2006")
	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	// render the summary page with the collected data
	render.Template(w, r, "reservation-summary.page.tmpl", &models.TemplateData{
		Data:      data,
		StringMap: stringMap,
	})
}

// ChooseRoom displays list of available rooms
func (m *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {
	// convert and store roomID
	roomID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Get the reservation from the session
	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		// Key not found or wrong type
		m.App.Session.Put(r.Context(), "error", "Reservation data not found in session")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// set the RoomID from the selected link
	res.RoomID = roomID

	// put the updated reservation back into the session
	m.App.Session.Put(r.Context(), "reservation", res)

	// redirect to the reservation page
	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)

}

// BookRoom takes URL parameters, builds a sessional variable, and redirects user to make res screen
func (m *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {
	roomID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	sd := r.URL.Query().Get("s")
	ed := r.URL.Query().Get("e")

	layout := "01/02/2006"
	startDate, err := time.Parse(layout, sd)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	endDate, err := time.Parse(layout, ed)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	var res models.Reservation

	room, err := m.DB.GetRoomByID(roomID)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	res.Room.RoomName = room.RoomName

	res.RoomID = roomID
	res.StartDate = startDate
	res.EndDate = endDate

	m.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}
