package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/bensabler/milos-residence/internal/models"
)

// postData represents form data structure for POST request testing
type postData struct {
	key   string
	value string
}

// routeTest defines the structure for testing HTTP routes with expected status codes
type routeTest struct {
	name               string     // descriptive name for the test case
	url                string     // the URL path to test
	method             string     // HTTP method (GET, POST, etc.)
	params             []postData // form parameters for POST requests
	expectedStatusCode int        // expected HTTP status code
	expectedLocation   string     // expected redirect location (for redirects)
	expectedInBody     string     // expected content in response body
}

// TestHandlers validates that all main application routes return correct HTTP status codes.
// This test ensures basic route accessibility and proper handler registration.
func TestHandlers(t *testing.T) {
	// Define comprehensive test cases covering all public routes
	tests := []routeTest{
		// Main informational pages - should all return 200 OK
		{"home", "/", "GET", []postData{}, http.StatusOK, "", ""},
		{"about", "/about", "GET", []postData{}, http.StatusOK, "", ""},
		{"photos", "/photos", "GET", []postData{}, http.StatusOK, "", ""},

		// Room detail pages - individual snooze spot information
		{"golden-haybeam-loft", "/golden-haybeam-loft", "GET", []postData{}, http.StatusOK, "", ""},
		{"window-perch-theater", "/window-perch-theater", "GET", []postData{}, http.StatusOK, "", ""},
		{"laundry-basket-nook", "/laundry-basket-nook", "GET", []postData{}, http.StatusOK, "", ""},

		// Availability and booking workflow pages
		{"search-availability", "/search-availability", "GET", []postData{}, http.StatusOK, "", ""},
		{"contact", "/contact", "GET", []postData{}, http.StatusOK, "", ""},

		// Note: Authentication routes tested separately due to middleware requirements
	}

	// Initialize the HTTP routes for testing
	routes := getRoutes()
	// Create test server with TLS support for realistic testing environment
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	// Execute each test case systematically
	for _, e := range tests {
		t.Run(e.name, func(t *testing.T) {
			var resp *http.Response
			var err error

			// Handle different HTTP methods appropriately
			if e.method == "GET" {
				// Perform GET request to the test server
				resp, err = ts.Client().Get(ts.URL + e.url)
			} else {
				// Prepare form data for POST requests
				values := url.Values{}
				for _, x := range e.params {
					values.Add(x.key, x.value)
				}
				// Execute POST request with form data
				resp, err = ts.Client().PostForm(ts.URL+e.url, values)
			}

			// Verify request execution succeeded
			if err != nil {
				t.Log(err)
				t.Fatal(err)
			}

			// Validate response status code matches expectation
			if resp.StatusCode != e.expectedStatusCode {
				t.Errorf("for %s expected %d but got %d", e.name, e.expectedStatusCode, resp.StatusCode)
			}

			// Check redirect location if specified in test case
			if e.expectedLocation != "" {
				actualLoc, _ := resp.Location()
				if actualLoc.String() != e.expectedLocation {
					t.Errorf("for %s expected location %s but got %s", e.name, e.expectedLocation, actualLoc.String())
				}
			}

			// Verify expected content is present in response body
			if e.expectedInBody != "" {
				// Read response body content
				buf := make([]byte, 1024)
				n, _ := resp.Body.Read(buf)
				body := string(buf[:n])

				if !strings.Contains(body, e.expectedInBody) {
					t.Errorf("for %s expected to find %s in body", e.name, e.expectedInBody)
				}
			}
		})
	}
}

// TestRepository_MakeReservation validates the reservation display handler functionality.
// This test verifies proper session handling, reservation data display, and error conditions.
func TestRepository_MakeReservation(t *testing.T) {
	// Define test cases for different reservation scenarios
	tests := []struct {
		name           string
		reservation    *models.Reservation
		setupSession   bool
		expectedStatus int
	}{
		{
			name: "valid reservation in session",
			reservation: &models.Reservation{
				RoomID: 1,
				Room: models.Room{
					ID:       1,
					RoomName: "Golden Haybeam Loft",
				},
			},
			setupSession:   true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "no reservation in session",
			reservation:    nil,
			setupSession:   false,
			expectedStatus: http.StatusSeeOther,
		},
		{
			name: "invalid room ID in reservation",
			reservation: &models.Reservation{
				RoomID: 100, // Non-existent room ID that will trigger error
				Room: models.Room{
					ID:       100,
					RoomName: "Non-existent Room",
				},
			},
			setupSession:   true,
			expectedStatus: http.StatusSeeOther,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create HTTP request for the make-reservation endpoint
			req, _ := http.NewRequest("GET", "/make-reservation", nil)

			// Set up request context with session support
			ctx := getCtx(req)
			req = req.WithContext(ctx)

			// Create response recorder to capture handler output
			rr := httptest.NewRecorder()

			// Configure session state based on test case requirements
			if tt.setupSession && tt.reservation != nil {
				// Store reservation data in session for positive test cases
				session.Put(ctx, "reservation", *tt.reservation)
			}

			// Execute the handler function
			handler := http.HandlerFunc(Repo.MakeReservation)
			handler.ServeHTTP(rr, req)

			// Validate response status code matches expectation
			if rr.Code != tt.expectedStatus {
				t.Errorf("MakeReservation handler returned wrong response code: got %d, wanted %d",
					rr.Code, tt.expectedStatus)
			}
		})
	}
}

// TestRepository_PostReservation validates the reservation submission processing.
// This comprehensive test covers form validation, database operations, and error handling.
func TestRepository_PostReservation(t *testing.T) {
	// Define test cases covering various submission scenarios
	tests := []struct {
		name           string
		formData       map[string]string
		expectedStatus int
		description    string
	}{
		{
			name: "successful reservation",
			formData: map[string]string{
				"start_date": "01/01/2100",
				"end_date":   "01/02/2100",
				"first_name": "John",
				"last_name":  "Smith",
				"email":      "john@smith.com",
				"phone":      "1234567891",
				"room_id":    "1",
			},
			expectedStatus: http.StatusSeeOther,
			description:    "Valid form data should redirect to confirmation",
		},
		{
			name: "invalid start date",
			formData: map[string]string{
				"start_date": "invalid",
				"end_date":   "01/02/2100",
				"first_name": "John",
				"last_name":  "Smith",
				"email":      "john@smith.com",
				"phone":      "1234567891",
				"room_id":    "1",
			},
			expectedStatus: http.StatusSeeOther,
			description:    "Invalid date format should redirect with error",
		},
		{
			name: "validation failure - short first name",
			formData: map[string]string{
				"start_date": "01/01/2100",
				"end_date":   "01/02/2100",
				"first_name": "J",
				"last_name":  "Smith",
				"email":      "john@smith.com",
				"phone":      "1234567891",
				"room_id":    "1",
			},
			expectedStatus: http.StatusOK,
			description:    "Form validation errors should re-render form",
		},
		{
			name: "database error simulation",
			formData: map[string]string{
				"start_date": "01/01/2100",
				"end_date":   "01/02/2100",
				"first_name": "John",
				"last_name":  "Smith",
				"email":      "john@smith.com",
				"phone":      "1234567891",
				"room_id":    "2", // Room ID 2 triggers error in test repo
			},
			expectedStatus: http.StatusSeeOther,
			description:    "Database errors should redirect with error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Construct form data string from test parameters
			var reqBody strings.Builder
			first := true
			for key, value := range tt.formData {
				if !first {
					reqBody.WriteString("&")
				}
				reqBody.WriteString(fmt.Sprintf("%s=%s", key, value))
				first = false
			}

			// Create POST request with form data
			req, _ := http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody.String()))

			// Set up request context and headers
			ctx := getCtx(req)
			req = req.WithContext(ctx)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			// Create response recorder for capturing handler output
			rr := httptest.NewRecorder()

			// Execute the PostReservation handler
			handler := http.HandlerFunc(Repo.PostReservation)
			handler.ServeHTTP(rr, req)

			// Validate response status matches expected outcome
			if rr.Code != tt.expectedStatus {
				t.Errorf("PostReservation %s: expected status %d, got %d - %s",
					tt.name, tt.expectedStatus, rr.Code, tt.description)
			}
		})
	}
}

// TestRepository_AvailabilityJSON validates the JSON API endpoint for room availability.
// This test ensures proper JSON response format and availability checking logic.
func TestRepository_AvailabilityJSON(t *testing.T) {
	// Define test cases for different availability scenarios
	tests := []struct {
		name         string
		formData     map[string]string
		expectedOK   bool
		expectedCode int
	}{
		{
			name: "valid availability request",
			formData: map[string]string{
				"start":   "01/01/2100",
				"end":     "01/02/2100",
				"room_id": "1",
			},
			expectedOK:   false, // Test repo returns false for availability
			expectedCode: http.StatusOK,
		},
		{
			name: "invalid date format",
			formData: map[string]string{
				"start":   "invalid",
				"end":     "01/02/2100",
				"room_id": "1",
			},
			expectedOK:   false,
			expectedCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare form data for the request
			formData := url.Values{}
			for key, value := range tt.formData {
				formData.Add(key, value)
			}

			// Create POST request with form data
			req, _ := http.NewRequest("POST", "/search-availability-json",
				strings.NewReader(formData.Encode()))

			// Configure request headers and context
			ctx := getCtx(req)
			req = req.WithContext(ctx)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			// Set up response recorder
			rr := httptest.NewRecorder()

			// Execute the availability JSON handler
			handler := http.HandlerFunc(Repo.AvailabilityJSON)
			handler.ServeHTTP(rr, req)

			// Verify response status code
			if rr.Code != tt.expectedCode {
				t.Errorf("AvailabilityJSON %s: expected status %d, got %d",
					tt.name, tt.expectedCode, rr.Code)
			}

			// Parse and validate JSON response structure
			var response jsonResponse
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			if err != nil {
				t.Errorf("Failed to parse JSON response: %v", err)
			}

			// Validate response contains expected availability result
			if response.OK != tt.expectedOK {
				t.Errorf("AvailabilityJSON %s: expected OK=%v, got OK=%v",
					tt.name, tt.expectedOK, response.OK)
			}
		})
	}
}

// TestRepository_PostAvailability validates the availability search form submission.
// This test checks date processing, database queries, and result handling.
func TestRepository_PostAvailability(t *testing.T) {
	// Define test scenarios for availability search
	tests := []struct {
		name           string
		start          string
		end            string
		expectedStatus int
		expectRedirect bool
	}{
		{
			name:           "valid date range",
			start:          "01/01/2100",
			end:            "01/02/2100",
			expectedStatus: http.StatusSeeOther, // No rooms available redirects with error
			expectRedirect: true,
		},
		{
			name:           "invalid start date",
			start:          "invalid",
			end:            "01/02/2100",
			expectedStatus: http.StatusSeeOther,
			expectRedirect: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create form data with date parameters
			formData := url.Values{}
			formData.Add("start", tt.start)
			formData.Add("end", tt.end)

			// Construct POST request
			req, _ := http.NewRequest("POST", "/search-availability",
				strings.NewReader(formData.Encode()))

			// Set up request context and headers
			ctx := getCtx(req)
			req = req.WithContext(ctx)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute the availability search handler
			handler := http.HandlerFunc(Repo.PostAvailability)
			handler.ServeHTTP(rr, req)

			// Validate response status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("PostAvailability %s: expected status %d, got %d",
					tt.name, tt.expectedStatus, rr.Code)
			}
		})
	}
}

// TestRepository_ReservationSummary validates the reservation confirmation display.
// This test ensures proper session data retrieval and summary page rendering.
func TestRepository_ReservationSummary(t *testing.T) {
	tests := []struct {
		name           string
		reservation    *models.Reservation
		setupSession   bool
		expectedStatus int
	}{
		{
			name: "valid reservation summary",
			reservation: &models.Reservation{
				ID:        1,
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
				RoomID:    1,
				StartDate: time.Now(),
				EndDate:   time.Now().AddDate(0, 0, 2),
			},
			setupSession:   true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "no reservation in session",
			reservation:    nil,
			setupSession:   false,
			expectedStatus: http.StatusSeeOther,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request for reservation summary page
			req, _ := http.NewRequest("GET", "/reservation-summary", nil)

			// Set up request context with session
			ctx := getCtx(req)
			req = req.WithContext(ctx)

			// Configure session data based on test requirements
			if tt.setupSession && tt.reservation != nil {
				session.Put(ctx, "reservation", *tt.reservation)
			}

			// Set up response recorder
			rr := httptest.NewRecorder()

			// Execute reservation summary handler
			handler := http.HandlerFunc(Repo.ReservationSummary)
			handler.ServeHTTP(rr, req)

			// Verify response status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("ReservationSummary %s: expected status %d, got %d",
					tt.name, tt.expectedStatus, rr.Code)
			}
		})
	}
}

// TestRepository_ChooseRoom validates room selection functionality.
// This test verifies URL parameter parsing and session management.
func TestRepository_ChooseRoom(t *testing.T) {
	tests := []struct {
		name           string
		roomID         string
		setupSession   bool
		expectedStatus int
	}{
		{
			name:           "valid room selection",
			roomID:         "1",
			setupSession:   true,
			expectedStatus: http.StatusSeeOther,
		},
		{
			name:           "invalid room ID",
			roomID:         "invalid",
			setupSession:   true,
			expectedStatus: http.StatusSeeOther,
		},
		{
			name:           "no session reservation",
			roomID:         "1",
			setupSession:   false,
			expectedStatus: http.StatusSeeOther,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request with proper RequestURI format for handler
			req, _ := http.NewRequest("GET", "/choose-room/"+tt.roomID, nil)
			req.RequestURI = "/choose-room/" + tt.roomID // Set RequestURI explicitly for handler parsing

			// Set up request context
			ctx := getCtx(req)
			req = req.WithContext(ctx)

			// Configure session with reservation data if needed
			if tt.setupSession {
				reservation := models.Reservation{RoomID: 1}
				session.Put(ctx, "reservation", reservation)
			}

			// Set up response recorder
			rr := httptest.NewRecorder()

			// Execute choose room handler
			handler := http.HandlerFunc(Repo.ChooseRoom)
			handler.ServeHTTP(rr, req)

			// Validate response status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("ChooseRoom %s: expected status %d, got %d",
					tt.name, tt.expectedStatus, rr.Code)
			}
		})
	}
}

// TestRepository_BookRoom validates direct room booking from URL parameters.
// This test checks query parameter processing and reservation setup.
func TestRepository_BookRoom(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "valid booking parameters",
			queryParams:    "?id=1&s=01/01/2100&e=01/02/2100",
			expectedStatus: http.StatusSeeOther,
		},
		{
			name:           "missing parameters",
			queryParams:    "?id=1",
			expectedStatus: http.StatusSeeOther, // Should handle gracefully
		},
		{
			name:           "invalid room ID",
			queryParams:    "?id=100&s=01/01/2100&e=01/02/2100",
			expectedStatus: http.StatusSeeOther,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request with query parameters
			req, _ := http.NewRequest("GET", "/book-room"+tt.queryParams, nil)

			// Set up request context
			ctx := getCtx(req)
			req = req.WithContext(ctx)

			// Set up response recorder
			rr := httptest.NewRecorder()

			// Execute book room handler
			handler := http.HandlerFunc(Repo.BookRoom)
			handler.ServeHTTP(rr, req)

			// Validate response status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("BookRoom %s: expected status %d, got %d",
					tt.name, tt.expectedStatus, rr.Code)
			}
		})
	}
}

// TestRepository_ShowLogin validates the login page display functionality.
// This test ensures proper form rendering and CSRF token inclusion.
func TestRepository_ShowLogin(t *testing.T) {
	// Create request for login page
	req, _ := http.NewRequest("GET", "/user/login", nil)

	// Set up request context
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	// Set up response recorder
	rr := httptest.NewRecorder()

	// Execute show login handler
	handler := http.HandlerFunc(Repo.ShowLogin)
	handler.ServeHTTP(rr, req)

	// Validate successful page rendering
	if rr.Code != http.StatusOK {
		t.Errorf("ShowLogin handler returned wrong response code: got %d, wanted %d",
			rr.Code, http.StatusOK)
	}
}

// TestRepository_LoginRouteIntegration validates login page through routing.
// This test verifies the login route is properly configured in the router.
func TestRepository_LoginRouteIntegration(t *testing.T) {
	// Get the configured routes
	routes := getRoutes()
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	// Test login page access through router
	resp, err := ts.Client().Get(ts.URL + "/user/login")
	if err != nil {
		t.Fatal(err)
	}

	// Validate login page is accessible
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Login page via router: expected %d but got %d", http.StatusOK, resp.StatusCode)
	}
}

// TestRepository_PostShowLogin validates user authentication processing.
// This test covers login form submission, validation, and authentication flow.
func TestRepository_PostShowLogin(t *testing.T) {
	tests := []struct {
		name           string
		email          string
		password       string
		expectedStatus int
	}{
		{
			name:           "successful login",
			email:          "test@example.com",
			password:       "password",
			expectedStatus: http.StatusSeeOther,
		},
		{
			name:           "missing email",
			email:          "",
			password:       "password",
			expectedStatus: http.StatusOK, // Re-render form with validation errors
		},
		{
			name:           "invalid email format",
			email:          "invalid-email",
			password:       "password",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create form data for login submission
			formData := url.Values{}
			formData.Add("email", tt.email)
			formData.Add("password", tt.password)

			// Create POST request with login credentials
			req, _ := http.NewRequest("POST", "/user/login",
				strings.NewReader(formData.Encode()))

			// Configure request headers and context
			ctx := getCtx(req)
			req = req.WithContext(ctx)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			// Set up response recorder
			rr := httptest.NewRecorder()

			// Execute login processing handler
			handler := http.HandlerFunc(Repo.PostShowLogin)
			handler.ServeHTTP(rr, req)

			// Validate response status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("PostShowLogin %s: expected status %d, got %d",
					tt.name, tt.expectedStatus, rr.Code)
			}
		})
	}
}

// TestRepository_Logout validates user session termination functionality.
// This test ensures proper session cleanup and redirect behavior.
func TestRepository_Logout(t *testing.T) {
	// Create logout request
	req, _ := http.NewRequest("GET", "/user/logout", nil)

	// Set up authenticated session context
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	// Set up user session to simulate authenticated state
	session.Put(ctx, "user_id", 1)

	// Set up response recorder
	rr := httptest.NewRecorder()

	// Execute logout handler
	handler := http.HandlerFunc(Repo.Logout)
	handler.ServeHTTP(rr, req)

	// Validate redirect to login page
	if rr.Code != http.StatusSeeOther {
		t.Errorf("Logout handler returned wrong response code: got %d, wanted %d",
			rr.Code, http.StatusSeeOther)
	}
}

// getCtx creates a request context with session support for testing.
// This helper function simplifies session-based test setup.
func getCtx(req *http.Request) context.Context {
	// Load session from request headers (simulated for testing)
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		log.Println(err)
	}
	return ctx
}

// TestRepository_StaticRoomPages validates individual room detail page handlers.
// This test ensures all room-specific pages render correctly.
func TestRepository_StaticRoomPages(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		url     string
	}{
		{"golden haybeam loft", Repo.GoldenHaybeamLoft, "/golden-haybeam-loft"},
		{"window perch theater", Repo.WindowPerchTheater, "/window-perch-theater"},
		{"laundry basket nook", Repo.LaundryBasketNook, "/laundry-basket-nook"},
		{"about page", Repo.About, "/about"},
		{"photos page", Repo.Photos, "/photos"},
		{"contact page", Repo.Contact, "/contact"},
		{"home page", Repo.Home, "/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request for the specific page
			req, _ := http.NewRequest("GET", tt.url, nil)

			// Set up request context
			ctx := getCtx(req)
			req = req.WithContext(ctx)

			// Set up response recorder
			rr := httptest.NewRecorder()

			// Execute the page handler
			tt.handler.ServeHTTP(rr, req)

			// Validate successful page rendering
			if rr.Code != http.StatusOK {
				t.Errorf("%s handler returned wrong response code: got %d, wanted %d",
					tt.name, rr.Code, http.StatusOK)
			}
		})
	}
}

// TestRepository_AdminShowReservation validates reservation detail display in admin interface.
// This test covers URL parsing, database operations, and template rendering for admin reservation views.
func TestRepository_AdminShowReservation(t *testing.T) {
	tests := []struct {
		name           string
		requestURI     string
		queryParams    string
		reservationID  int
		expectedStatus int
		shouldError    bool
	}{
		{
			name:           "valid reservation display",
			requestURI:     "/admin/reservations/new/1/show",
			queryParams:    "?y=2025&m=12",
			reservationID:  1,
			expectedStatus: http.StatusOK,
			shouldError:    false,
		},

		{
			name:           "invalid reservation ID",
			requestURI:     "/admin/reservations/new/invalid/show",
			queryParams:    "",
			reservationID:  0,
			expectedStatus: http.StatusInternalServerError,
			shouldError:    true,
		},
		{
			name:           "database error - reservation not found",
			requestURI:     "/admin/reservations/new/999/show",
			queryParams:    "",
			reservationID:  999,
			expectedStatus: http.StatusOK, // Test repo returns empty reservation, no error
			shouldError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request with specific RequestURI for URL parsing
			req, _ := http.NewRequest("GET", tt.requestURI+tt.queryParams, nil)
			req.RequestURI = tt.requestURI

			// Set up request context
			ctx := getCtx(req)
			req = req.WithContext(ctx)

			// Set up response recorder
			rr := httptest.NewRecorder()

			// Execute admin show reservation handler
			handler := http.HandlerFunc(Repo.AdminShowReservation)
			handler.ServeHTTP(rr, req)

			// Validate response status matches expectation
			if rr.Code != tt.expectedStatus {
				t.Errorf("AdminShowReservation %s: expected status %d, got %d",
					tt.name, tt.expectedStatus, rr.Code)
			}
		})
	}
}

// TestRepository_AdminPostShowReservation validates reservation update processing in admin interface.
// This test covers form processing, database updates, and redirect logic for admin reservation edits.
func TestRepository_AdminPostShowReservation(t *testing.T) {
	tests := []struct {
		name           string
		requestURI     string
		formData       map[string]string
		reservationID  int
		expectedStatus int
	}{
		{
			name:       "successful reservation update",
			requestURI: "/admin/reservations/new/1/show",
			formData: map[string]string{
				"first_name": "UpdatedJohn",
				"last_name":  "UpdatedDoe",
				"email":      "updated@example.com",
				"phone":      "1234567890",
				"month":      "",
				"year":       "",
			},
			reservationID:  1,
			expectedStatus: http.StatusSeeOther,
		},
		{
			name:       "successful update with calendar redirect",
			requestURI: "/admin/reservations/cal/1/show",
			formData: map[string]string{
				"first_name": "CalendarJohn",
				"last_name":  "CalendarDoe",
				"email":      "calendar@example.com",
				"phone":      "1234567890",
				"month":      "12",
				"year":       "2025",
			},
			reservationID:  1,
			expectedStatus: http.StatusSeeOther,
		},

		{
			name:       "invalid reservation ID",
			requestURI: "/admin/reservations/new/invalid/show",
			formData: map[string]string{
				"first_name": "Test",
			},
			reservationID:  0,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:       "database error - reservation not found",
			requestURI: "/admin/reservations/new/999/show",
			formData: map[string]string{
				"first_name": "Test",
			},
			reservationID:  999,
			expectedStatus: http.StatusSeeOther, // Test repo returns empty reservation, redirects
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare form data for POST request
			formData := url.Values{}
			for key, value := range tt.formData {
				formData.Add(key, value)
			}

			// Create POST request with form data
			req, _ := http.NewRequest("POST", tt.requestURI, strings.NewReader(formData.Encode()))
			req.RequestURI = tt.requestURI
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			// Set up request context
			ctx := getCtx(req)
			req = req.WithContext(ctx)

			// Set up response recorder
			rr := httptest.NewRecorder()

			// Execute admin post show reservation handler
			handler := http.HandlerFunc(Repo.AdminPostShowReservation)
			handler.ServeHTTP(rr, req)

			// Validate response status matches expectation
			if rr.Code != tt.expectedStatus {
				t.Errorf("AdminPostShowReservation %s: expected status %d, got %d",
					tt.name, tt.expectedStatus, rr.Code)
			}
		})
	}
}

// TestRepository_PostReservation_CompleteCoverage provides comprehensive testing for reservation processing.
// This test specifically targets the uncovered success paths and error scenarios.
func TestRepository_PostReservation_CompleteCoverage(t *testing.T) {
	tests := []struct {
		name           string
		formData       map[string]string
		expectedStatus int
		description    string
	}{
		{
			name:     "form parsing error - simulate malformed request",
			formData: map[string]string{
				// Empty form data to potentially trigger parsing issues
			},
			expectedStatus: http.StatusSeeOther,
			description:    "Malformed form should redirect with error",
		},
		{
			name: "room not found during validation error handling",
			formData: map[string]string{
				"start_date": "01/01/2100",
				"end_date":   "01/02/2100",
				"first_name": "J", // Too short to trigger validation error
				"last_name":  "Smith",
				"email":      "john@smith.com",
				"phone":      "1234567891",
				"room_id":    "100", // Non-existent room for validation error path
			},
			expectedStatus: http.StatusSeeOther,
			description:    "Room not found during validation should redirect",
		},
		{
			name: "successful reservation with room restriction failure",
			formData: map[string]string{
				"start_date": "01/01/2100",
				"end_date":   "01/02/2100",
				"first_name": "John",
				"last_name":  "Smith",
				"email":      "john@smith.com",
				"phone":      "1234567891",
				"room_id":    "1000", // Room ID that triggers restriction error in test repo
			},
			expectedStatus: http.StatusSeeOther,
			description:    "Room restriction error should redirect with error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Construct form data string from test parameters
			var reqBody strings.Builder
			first := true
			for key, value := range tt.formData {
				if !first {
					reqBody.WriteString("&")
				}
				reqBody.WriteString(fmt.Sprintf("%s=%s", key, value))
				first = false
			}

			// Create POST request with form data
			req, _ := http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody.String()))

			// Set up request context and headers
			ctx := getCtx(req)
			req = req.WithContext(ctx)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			// Set up response recorder
			rr := httptest.NewRecorder()

			// Execute the PostReservation handler
			handler := http.HandlerFunc(Repo.PostReservation)
			handler.ServeHTTP(rr, req)

			// Validate response status matches expected outcome
			if rr.Code != tt.expectedStatus {
				t.Errorf("PostReservation_CompleteCoverage %s: expected status %d, got %d - %s",
					tt.name, tt.expectedStatus, rr.Code, tt.description)
			}
		})
	}
}
