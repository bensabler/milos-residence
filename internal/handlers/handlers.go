// Package handlers implements HTTP request handlers for Milo's Residence reservation system.
// It provides handlers for public pages, reservation management, availability checking,
// user authentication, and administrative functions using the Repository pattern.
package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
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

// Repo is the global repository instance used by all handlers.
// It provides access to application configuration and database operations
// throughout the handler functions.
var Repo *Repository

// Repository implements the Repository pattern for HTTP handlers.
// It encapsulates application configuration and database access,
// providing a clean interface for handling HTTP requests while
// maintaining separation of concerns between web layer and business logic.
type Repository struct {
	App *config.AppConfig       // Application configuration and shared services
	DB  repository.DatabaseRepo // Database operations interface
}

// NewRepo creates a new Repository instance with the provided application configuration
// and database connection. It initializes the repository with a PostgreSQL database
// implementation and returns a configured Repository ready for use by handlers.
//
// Parameters:
//   - a: Application configuration containing session management, logging, and other settings
//   - db: Database connection wrapper with connection pool and health checking
//
// Returns a configured Repository instance with PostgreSQL database access.
func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewPostgresRepo(db.SQL, a),
	}
}

// NewTestRepo creates a new Repository instance configured for testing.
// It uses a test database implementation that provides controlled responses
// for unit testing without requiring an actual database connection.
//
// Parameters:
//   - a: Application configuration for the test environment
//
// Returns a Repository instance with test database implementation.
func NewTestRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewTestingRepo(a),
	}
}

// NewHandlers sets the global repository instance for use by handler functions.
// This function should be called during application initialization to configure
// the handlers with the appropriate repository implementation.
//
// Parameters:
//   - r: Repository instance to be used by all handler functions
func NewHandlers(r *Repository) {
	Repo = r
}

// Home handles GET requests to the homepage route (/).
// It renders the home page template with basic template data,
// demonstrating a simple handler that calls a database method
// and renders a template without complex business logic.
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	m.DB.AllUsers()
	render.Template(w, r, "home.page.tmpl", &models.TemplateData{})
}

// About handles GET requests to the about page route (/about).
// It renders the about page template with empty template data,
// providing information about the residence and its amenities.
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "about.page.tmpl", &models.TemplateData{})
}

// Photos handles GET requests to the photos page route (/photos).
// It renders the photos page template displaying images of the residence
// and its various room offerings.
func (m *Repository) Photos(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "photos.page.tmpl", &models.TemplateData{})
}

// MakeReservation handles GET requests to display the reservation form.
// It retrieves reservation data from the user session, validates the room exists,
// and renders the reservation form with pre-populated data. If the session
// doesn't contain valid reservation data or the room cannot be found,
// it redirects to the home page with an error message.
func (m *Repository) MakeReservation(w http.ResponseWriter, r *http.Request) {
	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	room, err := m.DB.GetRoomByID(res.RoomID)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't find room!")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	res.Room.RoomName = room.RoomName

	m.App.Session.Put(r.Context(), "reservation", res)

	sd := res.StartDate.Format("01/02/2006")
	ed := res.EndDate.Format("01/02/2006")

	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	data := make(map[string]interface{})
	data["reservation"] = res

	td := &models.TemplateData{
		Data:      data,
		Form:      forms.New(nil),
		StringMap: stringMap,
	}

	render.Template(w, r, "make-reservation.page.tmpl", td)
}

// PostReservation handles POST requests to process reservation form submissions.
// It validates form data, creates a reservation record in the database,
// creates corresponding room restrictions, sends confirmation emails,
// and redirects to the reservation summary page. If validation fails,
// it re-renders the form with error messages.
//
// The handler performs the following steps:
// 1. Parses and validates form data including dates and guest information
// 2. Validates required fields and data formats using the forms package
// 3. Creates reservation and room restriction records in the database
// 4. Sends confirmation email to guest and notification email to staff
// 5. Stores reservation in session and redirects to summary page
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse form!")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	sd := r.Form.Get("start_date")
	ed := r.Form.Get("end_date")

	layout := "01/02/2006"

	startDate, err := time.Parse(layout, sd)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse start date")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	endDate, err := time.Parse(layout, ed)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't get parse end date")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	roomID, err := strconv.Atoi(r.Form.Get("room_id"))
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "invalid data!")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	reservation := models.Reservation{
		FirstName: r.Form.Get("first_name"),
		LastName:  r.Form.Get("last_name"),
		Phone:     r.Form.Get("phone"),
		Email:     r.Form.Get("email"),
		StartDate: startDate,
		EndDate:   endDate,
		RoomID:    roomID,
	}

	form := forms.New(r.PostForm)

	form.Required("first_name", "last_name", "email", "phone")
	form.MinLength("first_name", 3)
	form.IsEmail("email")

	if !form.Valid() {
		// Get room info for re-rendering the form
		room, err := m.DB.GetRoomByID(roomID)
		if err != nil {
			m.App.Session.Put(r.Context(), "error", "can't find room!")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		reservation.Room.RoomName = room.RoomName

		data := make(map[string]interface{})
		data["reservation"] = reservation

		sd := reservation.StartDate.Format("01/02/2006")
		ed := reservation.EndDate.Format("01/02/2006")

		stringMap := make(map[string]string)
		stringMap["start_date"] = sd
		stringMap["end_date"] = ed

		// Re-render the form with validation errors (200 status)
		render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
			Form:      form,
			Data:      data,
			StringMap: stringMap,
		})
		return
	}

	room, err := m.DB.GetRoomByID(roomID)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't find room!")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	reservation.Room.RoomName = room.RoomName

	newReservationID, err := m.DB.InsertReservation(reservation)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't insert reservation into database!")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	restriction := models.RoomRestriction{
		StartDate:     startDate,
		EndDate:       endDate,
		RoomID:        roomID,
		ReservationID: newReservationID,
		RestrictionID: 1,
	}

	err = m.DB.InsertRoomRestriction(restriction)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't insert room restriction!")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	htmlMessage := fmt.Sprintf(`
			<strong>Reservation Confirmation</strong><br>
			Dear %s, <br>
			This is to confirm your reservation from %s to %s.
	`, reservation.FirstName, reservation.StartDate.Format("01/02/2006"), reservation.EndDate.Format("01/02/2006"))

	msg := models.MailData{
		To:       reservation.Email,
		From:     "milo@milos-residence.com",
		Subject:  "Reservation Confirmation",
		Content:  htmlMessage,
		Template: "basic.html",
	}

	m.App.MailChan <- msg

	htmlMessage = fmt.Sprintf(`
			<strong>Reservation Notification</strong><br>
			A reservation has been made at Milo's Residence for the %s snooze spot from %s to %s.
	`, reservation.Room.RoomName, reservation.StartDate.Format("01/02/2006"), reservation.EndDate.Format("01/02/2006"))

	msg = models.MailData{
		To:      "you@there.com",
		From:    "milo@milos-residence.com",
		Subject: "Reservation Notification",
		Content: htmlMessage,
	}

	m.App.MailChan <- msg

	m.App.Session.Put(r.Context(), "reservation", reservation)

	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

// GoldenHaybeamLoft handles GET requests to display the Golden Haybeam Loft room page.
// It renders a detailed page showcasing this specific room with its amenities,
// photos, and booking options.
func (m *Repository) GoldenHaybeamLoft(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "golden-haybeam-loft.page.tmpl", &models.TemplateData{})
}

// WindowPerchTheater handles GET requests to display the Window Perch Theater room page.
// It renders a detailed page showcasing this specific room with its amenities,
// photos, and booking options.
func (m *Repository) WindowPerchTheater(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "window-perch-theater.page.tmpl", &models.TemplateData{})
}

// LaundryBasketNook handles GET requests to display the Laundry Basket Nook room page.
// It renders a detailed page showcasing this specific room with its amenities,
// photos, and booking options.
func (m *Repository) LaundryBasketNook(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "laundry-basket-nook.page.tmpl", &models.TemplateData{})
}

// Availability handles GET requests to display the availability search form.
// It renders a form where users can input their desired check-in and check-out
// dates to search for available rooms.
func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "search-availability.page.tmpl", &models.TemplateData{})
}

// PostAvailability handles POST requests to search for available rooms.
// It processes the search form, queries the database for available rooms
// during the specified date range, and either displays available rooms
// or redirects with an error message if no rooms are available.
//
// The handler:
// 1. Parses and validates the date inputs from the form
// 2. Queries the database for rooms available during the date range
// 3. If rooms are found, stores search criteria in session and shows room selection
// 4. If no rooms are available, redirects back to search with error message
func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse form!")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	start := r.Form.Get("start")
	end := r.Form.Get("end")

	layout := "01/02/2006"
	startDate, err := time.Parse(layout, start)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse start date!")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	endDate, err := time.Parse(layout, end)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse end date!")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	rooms, err := m.DB.SearchAvailabilityForAllRooms(startDate, endDate)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't get availability for rooms")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if len(rooms) == 0 {
		m.App.Session.Put(r.Context(), "error", "No availability")
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
		return
	}

	data := make(map[string]interface{})
	data["rooms"] = rooms

	res := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
	}

	m.App.Session.Put(r.Context(), "reservation", res)

	render.Template(w, r, "choose-room.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

// jsonResponse represents the structure of JSON responses returned by the AvailabilityJSON handler.
// It provides a consistent format for AJAX availability checking requests,
// including success status, error messages, and booking details.
type jsonResponse struct {
	OK        bool   `json:"ok"`         // Whether the room is available
	Message   string `json:"message"`    // Error message if not available
	RoomID    string `json:"room_id"`    // ID of the requested room
	StartDate string `json:"start_date"` // Formatted start date
	EndDate   string `json:"end_date"`   // Formatted end date
}

// AvailabilityJSON handles POST requests for AJAX availability checking.
// It processes room availability requests and returns JSON responses
// indicating whether the specified room is available for the given dates.
// This endpoint is used by frontend JavaScript to provide real-time
// availability feedback without page refreshes.
//
// The response includes:
// - ok: boolean indicating availability
// - message: error message if request failed
// - room_id, start_date, end_date: echoed back for frontend processing
func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		resp := jsonResponse{
			OK:      false,
			Message: "Internal server error",
		}

		out, _ := json.MarshalIndent(resp, "", "     ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	sd := r.Form.Get("start")
	ed := r.Form.Get("end")

	layout := "01/02/2006"
	startDate, _ := time.Parse(layout, sd)
	endDate, _ := time.Parse(layout, ed)

	roomID, _ := strconv.Atoi(r.Form.Get("room_id"))

	available, err := m.DB.SearchAvailabilityByDatesByRoomID(startDate, endDate, roomID)
	if err != nil {
		resp := jsonResponse{
			OK:      false,
			Message: "Error querying database",
		}

		out, _ := json.MarshalIndent(resp, "", "     ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	resp := jsonResponse{
		OK:        available,
		Message:   "",
		StartDate: sd,
		EndDate:   ed,
		RoomID:    strconv.Itoa(roomID),
	}

	out, _ := json.MarshalIndent(resp, "", "     ")

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

// Contact handles GET requests to display the contact form.
// It renders the contact page with an empty form ready for user input,
// allowing visitors to send messages to the residence administrators.
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "contact.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
	})
}

// PostContact handles POST requests to process contact form submissions.
// It validates the form data, performs spam detection using a honeypot field,
// sends email notifications to both the administration and the sender,
// and redirects with success or error messages.
//
// Security features:
// - Honeypot field detection to prevent automated spam submissions
// - Form validation for required fields and email format
// - Dual email notifications for proper message handling
func (m *Repository) PostContact(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse form!")
		http.Redirect(w, r, "/contact", http.StatusSeeOther)
		return
	}

	// Honeypot check should be early
	website := r.Form.Get("website")
	if website != "" {
		m.App.Session.Put(r.Context(), "error", "Spam detected")
		http.Redirect(w, r, "/contact", http.StatusSeeOther)
		return
	}

	name := r.Form.Get("name")
	email := r.Form.Get("email")
	topic := r.Form.Get("topic")
	message := r.Form.Get("message")

	form := forms.New(r.PostForm)
	form.Required("name", "email", "message")
	form.MinLength("name", 3)
	form.IsEmail("email")
	form.MinLength("message", 10)

	if !form.Valid() {
		render.Template(w, r, "contact.page.tmpl", &models.TemplateData{
			Form: form,
		})
		return
	}

	// Send email notification
	htmlMessage := fmt.Sprintf(`
		<strong>New Contact Form Message</strong><br><br>
		<strong>From:</strong> %s (%s)<br>
		<strong>Topic:</strong> %s<br><br>
		<strong>Message:</strong><br>
		%s
	`, name, email, topic, message)

	msg := models.MailData{
		To:       "admin@milosresidence.com", // Change to your email
		From:     email,
		Subject:  fmt.Sprintf("Contact Form: %s", topic),
		Content:  htmlMessage,
		Template: "basic.html",
	}

	m.App.MailChan <- msg

	// Send confirmation email to user
	confirmationMessage := fmt.Sprintf(`
		Hi %s,<br><br>
		Thank you for contacting Milo's Residence! We've received your message and will get back to you within 24 hours.<br><br>
		Best purrs,<br>
		The Milo's Residence Team
	`, name)

	confirmMsg := models.MailData{
		To:       email,
		From:     "hello@milosresidence.com",
		Subject:  "Thanks for contacting Milo's Residence",
		Content:  confirmationMessage,
		Template: "basic.html",
	}

	m.App.MailChan <- confirmMsg
	// If the honeypot field is filled, treat it as spam and do not process further
	if website != "" {
		m.App.Session.Put(r.Context(), "error", "Spam detected")
		http.Redirect(w, r, "/contact", http.StatusSeeOther)
		return
	}

	m.App.Session.Put(r.Context(), "flash", "Thank you for your message! We'll get back to you soon.")
	http.Redirect(w, r, "/contact", http.StatusSeeOther)
}

// ReservationSummary handles GET requests to display reservation confirmation details.
// It retrieves the completed reservation from the session, displays the summary
// information to the user, and removes the reservation data from the session
// to prevent reuse. If no reservation data exists in the session,
// it redirects to the home page with an error message.
func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	m.App.Session.Remove(r.Context(), "reservation")

	data := make(map[string]interface{})
	data["reservation"] = reservation

	sd := reservation.StartDate.Format("01/02/2006")
	ed := reservation.EndDate.Format("01/02/2006")
	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	render.Template(w, r, "reservation-summary.page.tmpl", &models.TemplateData{
		Data:      data,
		StringMap: stringMap,
	})
}

// ChooseRoom handles GET requests to select a specific room for reservation.
// It extracts the room ID from the URL path, validates the room exists,
// updates the reservation in the session with the selected room,
// and redirects to the reservation form. If the session doesn't contain
// valid reservation data or the URL is malformed, it redirects with an error.
func (m *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {
	exploded := strings.Split(r.RequestURI, "/")
	roomID, err := strconv.Atoi(exploded[2])
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "missing url parameter")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	res.RoomID = roomID

	m.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

// BookRoom handles GET requests to initiate room booking from external links.
// It extracts booking parameters (room ID, start date, end date) from URL query parameters,
// validates the room exists, creates a reservation object, stores it in the session,
// and redirects to the reservation form. This handler enables direct booking links
// from room pages or external sources.
func (m *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {
	roomID, _ := strconv.Atoi(r.URL.Query().Get("id"))

	sd := r.URL.Query().Get("s")
	ed := r.URL.Query().Get("e")

	layout := "01/02/2006"
	startDate, _ := time.Parse(layout, sd)
	endDate, _ := time.Parse(layout, ed)

	var res models.Reservation

	room, err := m.DB.GetRoomByID(roomID)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't get room from db!")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	res.Room.RoomName = room.RoomName
	res.RoomID = roomID
	res.StartDate = startDate
	res.EndDate = endDate

	m.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

// ShowLogin handles GET requests to display the login form.
// It renders the login page with an empty form for user authentication,
// allowing staff and administrators to access protected areas of the application.
func (m *Repository) ShowLogin(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "login.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
	})
}

// PostShowLogin handles POST requests to process user login attempts.
// It validates the login form, attempts to authenticate the user credentials
// against the database, creates a new session upon successful authentication,
// and redirects to the home page. If authentication fails, it re-displays
// the login form with error messages.
//
// Security features:
// - Session token renewal to prevent session fixation attacks
// - Credential validation against hashed passwords in database
// - Error logging for failed authentication attempts
func (m *Repository) PostShowLogin(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	form := forms.New(r.PostForm)
	form.Required("email", "password")
	form.IsEmail("email")

	if !form.Valid() {
		render.Template(w, r, "login.page.tmpl", &models.TemplateData{
			Form: form,
		})
		return
	}

	id, _, err := m.DB.Authenticate(email, password)
	if err != nil {
		log.Println(err)
		m.App.Session.Put(r.Context(), "error", "Invalid login credentials")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	m.App.Session.Put(r.Context(), "user_id", id)
	m.App.Session.Put(r.Context(), "flash", "Logged in successfully!")
	http.Redirect(w, r, "/", http.StatusSeeOther)

}

// Logout handles GET requests to log users out of the application.
// It destroys the current session, creates a new session token for security,
// and redirects to the login page. This ensures complete session cleanup
// and prevents unauthorized access to protected resources.
func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.Destroy(r.Context())
	_ = m.App.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

// AdminDashboard handles GET requests to display the administrative dashboard.
// It renders the main admin interface page providing access to reservation
// management, reports, and other administrative functions. This handler
// requires authentication and is protected by middleware.
func (m *Repository) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "admin-dashboard.page.tmpl", &models.TemplateData{})
}

// AdminAllReservations handles GET requests to display all reservations.
// It retrieves all reservations from the database and renders them in
// a table format for administrative review. If database access fails,
// it returns an internal server error response.
func (m *Repository) AdminAllReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := m.DB.AllReservations()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations

	render.Template(w, r, "admin-all-reservations.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

// AdminNewReservations handles GET requests to display unprocessed reservations.
// It retrieves all new (unprocessed) reservations from the database and
// renders them in a table format for administrative processing. This allows
// staff to review and handle new booking requests efficiently.
func (m *Repository) AdminNewReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := m.DB.AllNewReservations()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations

	render.Template(w, r, "admin-new-reservations.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

// AdminShowReservation handles GET requests to display detailed reservation information.
// It extracts the reservation ID from the URL path, retrieves the complete
// reservation details from the database, and renders a detailed view with
// editing capabilities. URL parameters for year and month are preserved
// for navigation context when coming from calendar views.
func (m *Repository) AdminShowReservation(w http.ResponseWriter, r *http.Request) {

	exploded := strings.Split(r.RequestURI, "/")

	// Add these debug lines to see what's happening
	log.Printf("RequestURI: %s", r.RequestURI)
	log.Printf("Exploded parts: %v", exploded)
	if len(exploded) > 4 {
		log.Printf("Trying to convert exploded[4]: '%s'", exploded[4])
	} else {
		log.Printf("Not enough URL parts. Length: %d", len(exploded))
		helpers.ServerError(w, errors.New("malformed admin reservation URL"))
		return

	}

	id, err := strconv.Atoi(exploded[4])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	src := exploded[3]
	stringMap := make(map[string]string)
	stringMap["src"] = src

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	stringMap["month"] = month
	stringMap["year"] = year

	res, err := m.DB.GetReservationByID(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservation"] = res

	render.Template(w, r, "admin-reservations-show.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		Form:      forms.New(nil),
	})
}

// AdminPostShowReservation handles POST requests to update reservation details.
// It processes form submissions from the reservation detail page, updates
// the reservation information in the database, and redirects back to the
// appropriate listing (calendar or reservation list) based on the source context.
// Navigation context is preserved through hidden form fields.
func (m *Repository) AdminPostShowReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	exploded := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(exploded[4])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	src := exploded[3]
	stringMap := make(map[string]string)
	stringMap["src"] = src

	res, err := m.DB.GetReservationByID(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	res.FirstName = r.Form.Get("first_name")
	res.LastName = r.Form.Get("last_name")
	res.Email = r.Form.Get("email")
	res.Phone = r.Form.Get("phone")

	err = m.DB.UpdateReservation(res)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	month := r.Form.Get("month")
	year := r.Form.Get("year")

	m.App.Session.Put(r.Context(), "flash", "Changes saved")

	if year == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}

}

// AdminReservationsCalendar handles GET requests to display the reservation calendar view.
// It renders a monthly calendar showing room availability, existing reservations,
// and owner-blocked dates. The calendar supports navigation between months
// via query parameters and provides visual indicators for different types
// of room restrictions.
//
// Features:
// - Monthly calendar view with room-by-room availability
// - Visual distinction between reservations and owner blocks
// - Month navigation with preserved state
// - Interactive editing of room blocks
// - Session storage of block maps for form processing
func (m *Repository) AdminReservationsCalendar(w http.ResponseWriter, r *http.Request) {
	now := time.Now()

	if r.URL.Query().Get("y") != "" {
		year, _ := strconv.Atoi(r.URL.Query().Get("y"))
		month, _ := strconv.Atoi(r.URL.Query().Get("m"))

		now = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	}

	data := make(map[string]interface{})
	data["now"] = now

	next := now.AddDate(0, 1, 0)
	last := now.AddDate(0, -1, 0)

	nextMonth := next.Format("01")
	nextMonthYear := next.Format("2006")

	lastMonth := last.Format("01")
	lastMonthYear := last.Format("2006")

	stringMap := make(map[string]string)
	stringMap["next_month"] = nextMonth
	stringMap["next_month_year"] = nextMonthYear
	stringMap["last_month"] = lastMonth
	stringMap["last_month_year"] = lastMonthYear

	stringMap["this_month"] = now.Format("01")
	stringMap["this_month_year"] = now.Format("2006")

	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	intMap := make(map[string]int)
	intMap["days_in_month"] = lastOfMonth.Day()

	rooms, err := m.DB.AllRooms()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data["rooms"] = rooms

	for _, x := range rooms {
		reservationMap := make(map[string]int)
		blockMap := make(map[string]int)

		for d := firstOfMonth; !d.After(lastOfMonth); d = d.AddDate(0, 0, 1) {
			reservationMap[d.Format("01/02/2006")] = 0
			blockMap[d.Format("01/02/2006")] = 0
		}

		restrictions, err := m.DB.GetRestrictionsForRoomByDate(x.ID, firstOfMonth, lastOfMonth)
		if err != nil {
			helpers.ServerError(w, err)
			return
		}

		for _, y := range restrictions {
			if y.ReservationID > 0 {
				for d := y.StartDate; !d.After(y.EndDate); d = d.AddDate(0, 0, 1) {
					reservationMap[d.Format("01/02/2006")] = y.ReservationID
				}
			} else {
				blockMap[y.StartDate.Format("01/02/2006")] = y.ID
			}
		}
		data[fmt.Sprintf("reservation_map_%d", x.ID)] = reservationMap
		data[fmt.Sprintf("block_map_%d", x.ID)] = blockMap

		m.App.Session.Put(r.Context(), fmt.Sprintf("block_map_%d", x.ID), blockMap)

	}

	render.Template(w, r, "admin-reservations-calendar.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		IntMap:    intMap,
	})
}

// AdminProcessReservation handles GET requests to mark reservations as processed.
// It extracts the reservation ID from URL parameters, updates the reservation
// status in the database, and redirects back to the appropriate listing view.
// The handler preserves navigation context for seamless user experience
// when working with large reservation lists.
func (m *Repository) AdminProcessReservation(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	src := chi.URLParam(r, "src")

	err := m.DB.UpdateProcessedForReservation(id, 1)
	if err != nil {
		log.Println(err)
	}

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	m.App.Session.Put(r.Context(), "flash", "Reservation marked as processed!")

	if year == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)

	}

}

// AdminDeleteReservation handles GET requests to delete reservations.
// It extracts the reservation ID from URL parameters, removes the reservation
// from the database, and redirects back to the appropriate listing view.
// The handler preserves navigation context and provides user feedback
// through flash messages.
func (m *Repository) AdminDeleteReservation(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	src := chi.URLParam(r, "src")

	_ = m.DB.DeleteReservation(id)

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	m.App.Session.Put(r.Context(), "flash", "Reservation deleted!")

	if year == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)

	}

}

// AdminPostReservationsCalendar handles POST requests to update room availability blocks.
// It processes form submissions from the calendar view, managing room blocks
// (owner-restricted dates) by adding new blocks and removing existing ones
// based on checkbox selections. The handler compares current form state
// with stored session data to determine which blocks to add or remove.
//
// Processing logic:
// 1. Retrieves all rooms and their current block states from session
// 2. Removes blocks that were unchecked (removed checkboxes)
// 3. Adds new blocks for checked dates (added checkboxes)
// 4. Redirects back to calendar view with success message
func (m *Repository) AdminPostReservationsCalendar(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	year, _ := strconv.Atoi(r.Form.Get("y"))
	month, _ := strconv.Atoi(r.Form.Get("m"))

	rooms, err := m.DB.AllRooms()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	form := forms.New(r.PostForm)

	for _, x := range rooms {
		curMap := m.App.Session.Get(r.Context(), fmt.Sprintf("block_map_%d", x.ID)).(map[string]int)
		for name, value := range curMap {
			if val, ok := curMap[name]; ok {
				if val > 0 {
					if !form.Has(fmt.Sprintf("remove_block_%d_%s", x.ID, name)) {
						err := m.DB.DeleteBlockByID(value)
						if err != nil {
							log.Println(err)
						}
					}
				}
			}
		}
	}

	for name, _ := range r.PostForm {
		if strings.HasPrefix(name, "add_block") {
			exploded := strings.Split(name, "_")
			roomID, _ := strconv.Atoi(exploded[2])
			t, _ := time.Parse("01/02/2006", exploded[3])

			err := m.DB.InsertBlockForRoom(roomID, t)
			if err != nil {
				log.Println(err)
			}
		}
	}

	m.App.Session.Put(r.Context(), "flash", "Changes Saved")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%d&m=%d", year, month), http.StatusSeeOther)

}
