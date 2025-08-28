package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bensabler/milos-residence/internal/config"
	"github.com/bensabler/milos-residence/internal/driver"
	"github.com/bensabler/milos-residence/internal/forms"
	"github.com/bensabler/milos-residence/internal/models"
	"github.com/bensabler/milos-residence/internal/render"
	"github.com/bensabler/milos-residence/internal/repository"
	"github.com/bensabler/milos-residence/internal/repository/dbrepo"
)

// Repo implements the Singleton pattern for global handler access.
// This global variable provides a single point of access to the configured
// handler repository, allowing the routing system to access handlers consistently.
// While globals are generally discouraged, this pattern is common in web frameworks
// for handler registration and provides a clean API for route definitions.
var Repo *Repository

// Repository implements the Repository pattern and Dependency Injection.
// This struct holds all the dependencies that HTTP handlers need to process requests,
// including application configuration and database access. By centralizing these
// dependencies, we achieve loose coupling and make testing easier through dependency injection.
//
// Design Pattern: Repository pattern - encapsulates data access logic
// Design Pattern: Dependency Injection - receives dependencies rather than creating them
type Repository struct {
	App *config.AppConfig       // Application-wide configuration and shared services
	DB  repository.DatabaseRepo // Database abstraction layer for data operations
}

// NewRepo implements the Factory pattern for Repository creation.
// This factory function creates a fully configured Repository instance with all
// necessary dependencies injected. The factory pattern ensures that Repository
// instances are created consistently with proper initialization.
//
// Design Pattern: Factory Method - creates configured Repository instances
// Design Pattern: Dependency Injection - injects external dependencies
// Parameters:
//
//	a: Application configuration containing loggers, session manager, etc.
//	db: Database connection pool for data operations
//
// Returns: A fully configured Repository ready for use by handlers
func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,                                 // Store application config for access to loggers, sessions, templates
		DB:  dbrepo.NewPostgresRepo(db.SQL, a), // Create PostgreSQL-specific repository implementation
	}
}

// NewTestRepo implements the Test Double pattern for Repository creation.
// This factory creates Repository instances configured for testing, using mock
// or stub implementations instead of real database connections. This enables
// fast, isolated unit tests without requiring a database.
//
// Design Pattern: Factory Method - creates test-specific Repository instances
// Design Pattern: Test Double - provides mock implementations for testing
// Parameters:
//
//	a: Application configuration for test environment
//
// Returns: Repository configured with test doubles instead of real dependencies
func NewTestRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,                        // Use test application config
		DB:  dbrepo.NewTestingRepo(a), // Use mock database implementation
	}
}

// NewHandlers implements the Singleton initialization pattern.
// This function sets the global Repo variable, enabling route handlers to be
// accessed via the global singleton. While this creates a global dependency,
// it provides a clean interface for route registration in web applications.
//
// Design Pattern: Singleton - ensures single global instance
// Parameters:
//
//	r: The Repository instance to make globally available
func NewHandlers(r *Repository) {
	// Set the global repository reference for use by routing system
	// This allows routes to be defined as handlers.Repo.HandlerName
	Repo = r
}

// Home implements the Controller pattern for the application landing page.
// This handler demonstrates the Model-View-Controller (MVC) pattern where the handler
// acts as the Controller, coordinating between data (Model) and presentation (View).
// It shows how to render templates with minimal data processing.
//
// Design Pattern: Controller (from MVC) - handles HTTP requests and coordinates response
// Design Pattern: Template Method - uses render.Template for consistent page rendering
// HTTP Method: GET
// Route: /
// Parameters:
//
//	w: HTTP response writer for sending data back to client
//	r: HTTP request containing client data and context
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	// Demonstrate database connectivity by calling a simple query
	// This serves as both a health check and demonstrates the Repository pattern usage
	m.DB.AllUsers()

	// Render the home page template using the Template Method pattern
	// Pass empty TemplateData since the home page doesn't require dynamic content
	// The render.Template function handles template caching, error handling, and response writing
	render.Template(w, r, "home.page.tmpl", &models.TemplateData{})
}

// About implements the Controller pattern for the about page.
// This handler shows how to serve static content pages that don't require
// complex data processing or database queries. It demonstrates the simplest
// form of the MVC Controller pattern.
//
// Design Pattern: Controller (from MVC) - minimal controller for static content
// HTTP Method: GET
// Route: /about
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	// Render static about page - no dynamic data processing needed
	// This demonstrates how simple pages still benefit from the template system
	render.Template(w, r, "about.page.tmpl", &models.TemplateData{})
}

// Photos implements the Controller pattern for the photo gallery page.
// Similar to About, this serves relatively static content but could be extended
// to include dynamic photo loading from a database or file system.
//
// Design Pattern: Controller (from MVC) - static content controller
// HTTP Method: GET
// Route: /photos
func (m *Repository) Photos(w http.ResponseWriter, r *http.Request) {
	// Serve the photo gallery page
	// Future enhancement: could load dynamic photo data from database
	render.Template(w, r, "photos.page.tmpl", &models.TemplateData{})
}

// MakeReservation implements the Controller pattern for displaying the reservation form.
// This handler demonstrates the Session State pattern by retrieving data from the user's
// session and using it to populate a form. It shows how to handle session-based workflows
// where data flows between multiple HTTP requests.
//
// Design Pattern: Controller (from MVC) - form display controller
// Design Pattern: Session State - retrieves reservation data from session
// Design Pattern: Error Handling - graceful handling of missing session data
// HTTP Method: GET
// Route: /make-reservation
func (m *Repository) MakeReservation(w http.ResponseWriter, r *http.Request) {
	// Retrieve reservation data from session using Session State pattern
	// The session allows us to maintain state across the stateless HTTP protocol
	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		// Session data is missing or corrupted - implement graceful error handling
		// Store error message in session for display on redirect target page
		m.App.Session.Put(r.Context(), "error", "can't get reservation from session")
		// Redirect to home page rather than showing a broken form
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Enrich reservation data with room information using Repository pattern
	// This demonstrates how controllers coordinate between different data sources
	room, err := m.DB.GetRoomByID(res.RoomID)
	if err != nil {
		// Database error occurred - implement consistent error handling pattern
		m.App.Session.Put(r.Context(), "error", "can't find room!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Update the reservation with room name for display
	// This shows how to merge data from different sources
	res.Room.RoomName = room.RoomName

	// Update session with enriched reservation data
	// This ensures subsequent requests have complete information
	m.App.Session.Put(r.Context(), "reservation", res)

	// Format dates for display using consistent date formatting
	// This demonstrates data transformation for presentation layer
	sd := res.StartDate.Format("01/02/2006") // MM/dd/yyyy format for US users
	ed := res.EndDate.Format("01/02/2006")

	// Prepare template data using the Data Transfer Object pattern
	// StringMap holds simple key-value pairs for template consumption
	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	// Prepare complex data objects for template
	// Data map holds structured objects that templates can iterate over
	data := make(map[string]interface{})
	data["reservation"] = res

	// Create template data container with all necessary information
	// This demonstrates the View Model pattern - data specifically shaped for the view
	td := &models.TemplateData{
		Data:      data,           // Complex objects for template logic
		Form:      forms.New(nil), // Empty form for validation error display
		StringMap: stringMap,      // Simple key-value pairs for display
	}

	// Render the reservation form with pre-populated data
	render.Template(w, r, "make-reservation.page.tmpl", td)
}

// PostReservation implements the Controller pattern for processing reservation form submissions.
// This handler demonstrates the Command pattern by processing a form submission that changes
// system state. It shows comprehensive form processing including validation, data persistence,
// and user feedback through redirects and session messages.
//
// Design Pattern: Controller (from MVC) - form processing controller
// Design Pattern: Command - processes state-changing operation
// Design Pattern: Session State - maintains data across redirects
// Design Pattern: Repository - persists data through abstraction layer
// HTTP Method: POST
// Route: /make-reservation
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse form!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	sd := r.Form.Get("start_date")
	ed := r.Form.Get("end_date")

	layout := "01/02/2006"

	startDate, err := time.Parse(layout, sd)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse start date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	endDate, err := time.Parse(layout, ed)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't get parse end date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	roomID, err := strconv.Atoi(r.Form.Get("room_id"))
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "invalid data!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
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

	// Initialize form validation using the Validation pattern
	// This creates a wrapper around the raw form data with validation capabilities
	form := forms.New(r.PostForm)

	// Apply validation rules using the Strategy pattern
	// Each validation method implements a different validation strategy
	form.Required("first_name", "last_name", "email", "phone") // Presence validation
	form.MinLength("first_name", 3)                            // Length validation
	form.IsEmail("email")                                      // Format validation

	// Check validation results and handle failures
	if !form.Valid() {
		// Validation failed - re-render form with errors instead of redirecting
		// This demonstrates the PRG (Post-Redirect-Get) pattern avoidance for validation errors

		// Prepare template data including the form with errors
		data := make(map[string]interface{})
		data["reservation"] = reservation

		http.Error(w, "my own error message", http.StatusSeeOther)

		// Re-render the form template with validation errors displayed
		// The form object contains both the submitted values and error messages
		render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
			Form: form, // Form with validation errors for display
			Data: data, // Reservation data to repopulate form fields
		})
		return
	}

	// Validation passed - persist reservation using Repository pattern
	// The repository abstracts away database implementation details
	newReservationID, err := m.DB.InsertReservation(reservation)
	if err != nil {
		// Session data is missing or corrupted - implement graceful error handling
		// Store error message in session for display on redirect target page
		m.App.Session.Put(r.Context(), "error", "can't insert reservation into database!")
		// Redirect to home page rather than showing a broken form
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
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
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Success! Store reservation in session for confirmation page
	// This implements the Flash Message pattern through session storage
	m.App.Session.Put(r.Context(), "reservation", reservation)

	// Redirect to confirmation page using PRG (Post-Redirect-Get) pattern
	// This prevents form resubmission if user refreshes the page
	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

// GoldenHaybeamLoft implements the Controller pattern for a themed room page.
// This handler serves static content for a specific room type, demonstrating
// how to create content-specific controllers while maintaining consistent
// template rendering patterns.
//
// Design Pattern: Controller (from MVC) - content-specific controller
// HTTP Method: GET
// Route: /golden-haybeam-loft
func (m *Repository) GoldenHaybeamLoft(w http.ResponseWriter, r *http.Request) {
	// Render the themed room page
	// This demonstrates how the same template pattern works for all content types
	render.Template(w, r, "golden-haybeam-loft.page.tmpl", &models.TemplateData{})
}

// WindowPerchTheater implements the Controller pattern for another themed room page.
// Similar to GoldenHaybeamLoft, this shows how to create multiple themed content
// controllers using the same underlying template infrastructure.
//
// Design Pattern: Controller (from MVC) - themed content controller
// HTTP Method: GET
// Route: /window-perch-theater
func (m *Repository) WindowPerchTheater(w http.ResponseWriter, r *http.Request) {
	// Render the window perch themed page
	render.Template(w, r, "window-perch-theater.page.tmpl", &models.TemplateData{})
}

// LaundryBasketNook implements the Controller pattern for the third themed room page.
// This completes the set of themed room controllers, demonstrating consistent
// patterns across similar content types.
//
// Design Pattern: Controller (from MVC) - themed content controller
// HTTP Method: GET
// Route: /laundry-basket-nook
func (m *Repository) LaundryBasketNook(w http.ResponseWriter, r *http.Request) {
	// Render the laundry basket themed page
	render.Template(w, r, "laundry-basket-nook.page.tmpl", &models.TemplateData{})
}

// Availability implements the Controller pattern for displaying the availability search form.
// This handler serves the initial search interface where users can input date ranges
// to check room availability. It demonstrates how to create search interfaces
// that lead to more complex processing workflows.
//
// Design Pattern: Controller (from MVC) - search interface controller
// HTTP Method: GET
// Route: /search-availability
func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	// Render the availability search form
	// This form will POST back to PostAvailability for processing
	render.Template(w, r, "search-availability.page.tmpl", &models.TemplateData{})
}

// PostAvailability renders the search availability page
func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse form!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Extract search parameters from form submission
	// These represent the user's desired check-in and check-out dates
	start := r.Form.Get("start") // Start date as string (MM/dd/yyyy format)
	end := r.Form.Get("end")     // End date as string (MM/dd/yyyy format)

	// Parse date strings into Go time.Time objects for database queries
	// This demonstrates input validation and data type conversion
	layout := "01/02/2006" // US date format template for parsing
	startDate, err := time.Parse(layout, start)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse start date!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	endDate, err := time.Parse(layout, end)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse end date!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Query database for available rooms using Repository pattern
	// This abstracts the complex SQL logic behind a clean interface
	rooms, err := m.DB.SearchAvailabilityForAllRooms(startDate, endDate)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't get availability for rooms")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Handle case where no rooms are available
	if len(rooms) == 0 {
		// No availability - use Flash Message pattern to inform user
		// Store error message in session for display after redirect
		m.App.Session.Put(r.Context(), "error", "No availability")
		// Redirect back to search form using PRG pattern
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
		return
	}

	// Rooms are available - prepare data for room selection page
	// Create data container for template consumption using Data Transfer Object pattern
	data := make(map[string]interface{})
	data["rooms"] = rooms // Available rooms list for template iteration

	// Create reservation object to store search parameters in session
	// This demonstrates how search results flow into the booking workflow
	res := models.Reservation{
		StartDate: startDate, // Store parsed start date for booking
		EndDate:   endDate,   // Store parsed end date for booking
	}

	// Store search results in session for subsequent booking steps
	// This implements the Session State pattern for multi-step workflows
	m.App.Session.Put(r.Context(), "reservation", res)

	// Render room selection page with available rooms
	// This demonstrates the View Model pattern - data shaped specifically for the view
	render.Template(w, r, "choose-room.page.tmpl", &models.TemplateData{
		Data: data, // Available rooms for user selection
	})
}

// jsonResponse implements the Data Transfer Object pattern for API responses.
// This struct defines the contract for JSON responses sent to AJAX clients,
// providing a consistent API interface for availability checking functionality.
//
// Design Pattern: Data Transfer Object - structured data for API communication
type jsonResponse struct {
	OK        bool   `json:"ok"`         // Success/failure flag for client logic
	Message   string `json:"message"`    // Human-readable status message
	RoomID    string `json:"room_id"`    // Room identifier for booking
	StartDate string `json:"start_date"` // Formatted start date string
	EndDate   string `json:"end_date"`   // Formatted end date string
}

// AvailabilityJSON handles request for availability and send JSON response
func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	// need to parse request body
	err := r.ParseForm()
	if err != nil {
		// can't parse form, so return appropriate json
		resp := jsonResponse{
			OK:      false,
			Message: "Internal server error",
		}

		out, _ := json.MarshalIndent(resp, "", "     ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	// Extract search parameters from AJAX form submission
	sd := r.Form.Get("start") // Start date string from client
	ed := r.Form.Get("end")   // End date string from client

	// Parse dates using same logic as PostAvailability
	// This demonstrates consistency between different interface types (HTML vs JSON)
	layout := "01/02/2006"
	startDate, _ := time.Parse(layout, sd) // Note: Error handling simplified for brevity
	endDate, _ := time.Parse(layout, ed)

	// Extract room ID and convert to integer for database query
	roomID, _ := strconv.Atoi(r.Form.Get("room_id"))

	available, err := m.DB.SearchAvailabilityByDatesByRoomID(startDate, endDate, roomID)
	if err != nil {
		// got a database error, so return appropriate json
		resp := jsonResponse{
			OK:      false,
			Message: "Error querying database",
		}

		out, _ := json.MarshalIndent(resp, "", "     ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	// Prepare JSON response using Data Transfer Object pattern
	// This provides consistent API responses for JavaScript consumption
	resp := jsonResponse{
		OK:        available,            // Boolean flag for client logic
		Message:   "",                   // Could contain error details
		StartDate: sd,                   // Echo back the requested dates
		EndDate:   ed,                   // Echo back the requested dates
		RoomID:    strconv.Itoa(roomID), // Convert back to string for JSON
	}

	out, _ := json.MarshalIndent(resp, "", "     ")

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

// Contact implements the Controller pattern for the contact page.
// This handler serves the contact form and information page, demonstrating
// how to create customer service interfaces within the application.
//
// Design Pattern: Controller (from MVC) - contact interface controller
// HTTP Method: GET
// Route: /contact
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	// Render the contact page with form and information
	render.Template(w, r, "contact.page.tmpl", &models.TemplateData{})
}

// ReservationSummary displays the reservation summary page
func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Clean up session data to prevent stale information
	// This implements security best practice of not retaining sensitive data longer than needed
	m.App.Session.Remove(r.Context(), "reservation")

	// Prepare confirmation data for template display
	// This demonstrates the View Model pattern - shaping data for presentation
	data := make(map[string]interface{})
	data["reservation"] = reservation

	// Format dates for display using consistent formatting
	// This shows data transformation for user-friendly presentation
	sd := reservation.StartDate.Format("01/02/2006")
	ed := reservation.EndDate.Format("01/02/2006")
	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	// Render confirmation page with reservation details
	// This concludes the booking workflow with positive user feedback
	render.Template(w, r, "reservation-summary.page.tmpl", &models.TemplateData{
		Data:      data,      // Reservation object for detail display
		StringMap: stringMap, // Formatted dates for easy template access
	})
}

func (m *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {
	exploded := strings.Split(r.RequestURI, "/")
	roomID, err := strconv.Atoi(exploded[2])
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "missing url parameter")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	res.RoomID = roomID

	m.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

// BookRoom implements the Controller pattern for direct booking from room pages.
// This handler processes bookings initiated directly from room detail pages,
// demonstrating how to handle URL query parameters and initialize session state
// for booking workflows that bypass the search process.
//
// Design Pattern: Controller (from MVC) - direct booking controller
// Design Pattern: Session State - initializes workflow state
// HTTP Method: GET
// Route: /book-room (with query parameters)
func (m *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {
	// Extract booking parameters from URL query string
	// This demonstrates how to handle bookings initiated from room detail pages
	roomID, _ := strconv.Atoi(r.URL.Query().Get("id"))

	// Extract date parameters from query string
	sd := r.URL.Query().Get("s") // Start date string
	ed := r.URL.Query().Get("e") // End date string

	layout := "01/02/2006"
	startDate, _ := time.Parse(layout, sd)
	endDate, _ := time.Parse(layout, ed)

	var res models.Reservation

	room, err := m.DB.GetRoomByID(roomID)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't get room from db!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Populate reservation with room and date information
	// This demonstrates domain object initialization from multiple sources
	res.Room.RoomName = room.RoomName // Room name for display
	res.RoomID = roomID               // Room ID for database operations
	res.StartDate = startDate         // Booking start date
	res.EndDate = endDate             // Booking end date

	// Store initialized reservation in session to start booking workflow
	// This demonstrates how different entry points converge on the same workflow
	m.App.Session.Put(r.Context(), "reservation", res)

	// Redirect to reservation form to continue booking process
	// This demonstrates workflow convergence - different entry points, same next step
	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}
