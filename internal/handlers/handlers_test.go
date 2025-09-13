// Package handlers provides HTTP request handlers for the Milo's Residence application.
// This file contains comprehensive tests for all handler functions, covering both
// successful operations and error conditions using a test repository implementation.
//
// The tests use a toggle-based approach where global variables in the dbrepo package
// can be set to force specific database errors, enabling thorough testing of error
// handling paths without requiring complex mocking or actual database failures.
package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/bensabler/milos-residence/internal/config"
	"github.com/bensabler/milos-residence/internal/driver"
	"github.com/bensabler/milos-residence/internal/models"
	"github.com/bensabler/milos-residence/internal/repository/dbrepo"
	"github.com/go-chi/chi/v5"
)

// sessionize attaches session context to a request for handler testing.
// This is required because handlers expect session data to be available
// and will panic if session context is missing from the request.
func sessionize(req *http.Request) *http.Request {
	ctx, _ := session.Load(req.Context(), req.Header.Get("X-Session"))
	return req.WithContext(ctx)
}

// newGET creates a GET request with session context attached.
// Use this helper instead of httptest.NewRequest directly to ensure
// proper session handling in handler tests.
func newGET(path string) *http.Request {
	return sessionize(httptest.NewRequest(http.MethodGet, path, nil))
}

// newPOSTForm creates a POST request with form data and session context attached.
// The form data is URL-encoded and the appropriate Content-Type header is set.
// Session context is attached to prevent handler panics.
func newPOSTForm(path string, form url.Values) *http.Request {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return sessionize(req)
}

// do executes a handler function and returns the response recorder.
// This centralizes handler execution and provides a consistent way to
// capture responses for testing assertions.
func do(h http.HandlerFunc, req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

// mustStatus fails the test if the HTTP status code doesn't match the expected value.
// This helper provides clear error messages when status assertions fail and
// includes t.Helper() to ensure stack traces point to the calling test function.
func mustStatus(t *testing.T, rr *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rr.Code != want {
		t.Fatalf("status: got %d, want %d", rr.Code, want)
	}
}

// mustRedirectContains asserts that the Location header contains a specific substring.
// This is useful for testing redirects where the exact URL may include query parameters
// or other dynamic components, but you need to verify the redirect destination.
func mustRedirectContains(t *testing.T, rr *httptest.ResponseRecorder, wantSub string) {
	t.Helper()
	loc := rr.Header().Get("Location")
	if !strings.Contains(loc, wantSub) {
		t.Fatalf("redirect: got %q, want contains %q", loc, wantSub)
	}
}

// toForm converts a string map to url.Values for easy form data creation.
// This helper simplifies test setup by allowing form data to be specified
// as a regular map[string]string in test cases.
func toForm(m map[string]string) url.Values {
	v := url.Values{}
	for k, val := range m {
		v.Set(k, val)
	}
	return v
}

// ptrBool returns a pointer to a bool value for table-driven tests.
// This is needed when test cases need to distinguish between false and nil
// for optional boolean assertions.
func ptrBool(b bool) *bool { return &b }

// TestNewRepo verifies that NewRepo constructor creates a repository with proper configuration.
// This test ensures the repository is correctly initialized with the provided application
// configuration and database connection, and that all required fields are set.
func TestNewRepo(t *testing.T) {
	app := &config.AppConfig{}
	d := &driver.DB{SQL: &sql.DB{}}

	r := NewRepo(app, d)

	if r == nil {
		t.Fatal("NewRepo returned nil")
	}
	if r.App != app {
		t.Errorf("NewRepo.App mismatch: got %p want %p", r.App, app)
	}
	if r.DB == nil {
		t.Error("NewRepo.DB is nil; expected a concrete DatabaseRepo")
	}
}

// TestNewHandlers verifies that NewHandlers properly sets the global Repo variable.
// The global Repo variable is used by all handler functions, so this test ensures
// the initialization function correctly assigns the provided repository instance.
func TestNewHandlers(t *testing.T) {
	orig := Repo
	t.Cleanup(func() { NewHandlers(orig) })

	app := &config.AppConfig{Session: session}
	r := NewTestRepo(app)
	NewHandlers(r)

	if Repo == nil {
		t.Fatal("Repo is nil after NewHandlers")
	}
	if Repo != r {
		t.Errorf("Repo != passed repository; got %p want %p", Repo, r)
	}
}

// TestRoutes_Smoke ensures that all public routes are registered and return HTTP 200 OK.
// This test provides basic confidence that the routing configuration is correct
// and that public pages can be accessed without authentication.
func TestRoutes_Smoke(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"home", "/"},
		{"about", "/about"},
		{"photos", "/photos"},
		{"search-availability", "/search-availability"},
		{"golden-haybeam-loft", "/golden-haybeam-loft"},
		{"window-perch-theater", "/window-perch-theater"},
		{"laundry-basket-nook", "/laundry-basket-nook"},
		{"contact", "/contact"},
	}

	ts := httptest.NewTLSServer(getRoutes())
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := ts.Client().Get(ts.URL + tt.path)
			if err != nil {
				t.Fatalf("GET %s error: %v", tt.path, err)
			}
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("GET %s: status %d want %d", tt.path, resp.StatusCode, http.StatusOK)
			}
		})
	}
}

// TestRepository_MakeReservation verifies the reservation form display handler.
// This handler requires reservation data in the session and performs room lookup
// to populate the form. The test covers success cases, missing session data,
// and invalid room ID scenarios.
func TestRepository_MakeReservation(t *testing.T) {
	tests := []struct {
		name       string
		seed       *models.Reservation
		wantStatus int
	}{
		{
			name: "reservation present",
			seed: &models.Reservation{
				RoomID: 1,
				Room:   models.Room{ID: 1, RoomName: "Golden Haybeam Loft"},
			},
			wantStatus: http.StatusOK,
		},
		{name: "missing reservation", seed: nil, wantStatus: http.StatusSeeOther},
		{
			name: "invalid room id",
			seed: &models.Reservation{
				RoomID: 100, // test repo returns error for IDs > 3
				Room:   models.Room{ID: 100},
			},
			wantStatus: http.StatusSeeOther,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := newGET("/make-reservation")
			if tc.seed != nil {
				session.Put(req.Context(), "reservation", *tc.seed)
			}
			rr := do(Repo.MakeReservation, req)
			mustStatus(t, rr, tc.wantStatus)
		})
	}
}

// TestRepository_PostReservation_ParseFormAndRoomLookupErrors tests edge cases in reservation processing.
// These tests cover malformed form data and room lookup failures that occur after
// form validation succeeds but before database operations.
func TestRepository_PostReservation_ParseFormAndRoomLookupErrors(t *testing.T) {
	// Test malformed form data handling
	req := httptest.NewRequest(http.MethodPost, "/make-reservation", strings.NewReader("%not-urlencoded"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = sessionize(req)
	rr := do(Repo.PostReservation, req)
	mustStatus(t, rr, http.StatusSeeOther)

	// Test room lookup failure after successful validation
	req = newPOSTForm("/make-reservation", toForm(map[string]string{
		"start_date": "01/01/2100",
		"end_date":   "01/02/2100",
		"first_name": "John",
		"last_name":  "Smith",
		"email":      "john@smith.com",
		"phone":      "1234567891",
		"room_id":    "100", // triggers GetRoomByID error in test repo
	}))
	rr = do(Repo.PostReservation, req)
	mustStatus(t, rr, http.StatusSeeOther)
}

// TestRepository_PostReservation covers the full reservation creation workflow.
// This test exercises form validation, database operations, and error handling
// for the complete reservation submission process. It tests both success paths
// and various failure scenarios including validation errors and database failures.
func TestRepository_PostReservation(t *testing.T) {
	tests := []struct {
		name       string
		form       map[string]string
		wantStatus int
	}{
		{
			name: "success",
			form: map[string]string{
				"start_date": "01/01/2100",
				"end_date":   "01/02/2100",
				"first_name": "John",
				"last_name":  "Smith",
				"email":      "john@smith.com",
				"phone":      "1234567891",
				"room_id":    "1",
			},
			wantStatus: http.StatusSeeOther,
		},
		{
			name: "invalid start date",
			form: map[string]string{
				"start_date": "invalid",
				"end_date":   "01/02/2100",
				"first_name": "John",
				"last_name":  "Smith",
				"email":      "john@smith.com",
				"phone":      "1234567891",
				"room_id":    "1",
			},
			wantStatus: http.StatusSeeOther,
		},
		{
			name: "validation failure (first name too short)",
			form: map[string]string{
				"start_date": "01/01/2100",
				"end_date":   "01/02/2100",
				"first_name": "J", // fails MinLength validation
				"last_name":  "Smith",
				"email":      "john@smith.com",
				"phone":      "1234567891",
				"room_id":    "1",
			},
			wantStatus: http.StatusOK, // re-renders form with errors
		},
		{
			name: "insert reservation database error",
			form: map[string]string{
				"start_date": "01/01/2100",
				"end_date":   "01/02/2100",
				"first_name": "John",
				"last_name":  "Smith",
				"email":      "john@smith.com",
				"phone":      "1234567891",
				"room_id":    "2", // triggers error in test repo
			},
			wantStatus: http.StatusSeeOther,
		},
		{
			name: "room restriction insert error",
			form: map[string]string{
				"start_date": "01/01/2100",
				"end_date":   "01/02/2100",
				"first_name": "John",
				"last_name":  "Smith",
				"email":      "john@smith.com",
				"phone":      "1234567891",
				"room_id":    "3", // triggers restriction error in test repo
			},
			wantStatus: http.StatusSeeOther,
		},
		{
			name: "invalid end date",
			form: map[string]string{
				"start_date": "01/01/2100",
				"end_date":   "not-a-date",
				"first_name": "John",
				"last_name":  "Smith",
				"email":      "john@smith.com",
				"phone":      "1234567891",
				"room_id":    "1",
			},
			wantStatus: http.StatusSeeOther,
		},
		{
			name: "invalid room_id (non-numeric)",
			form: map[string]string{
				"start_date": "01/01/2100",
				"end_date":   "01/02/2100",
				"first_name": "John",
				"last_name":  "Smith",
				"email":      "john@smith.com",
				"phone":      "1234567891",
				"room_id":    "x", // invalid integer conversion
			},
			wantStatus: http.StatusSeeOther,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := newPOSTForm("/make-reservation", toForm(tc.form))
			rr := do(Repo.PostReservation, req)
			mustStatus(t, rr, tc.wantStatus)
		})
	}
}

// TestRepository_ReservationSummary verifies the reservation confirmation page.
// This handler displays completed reservation details and requires reservation
// data to be present in the session. The test covers both successful display
// and missing session data scenarios.
func TestRepository_ReservationSummary(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name       string
		seed       *models.Reservation
		wantStatus int
	}{
		{
			name: "reservation summary displays correctly",
			seed: &models.Reservation{
				ID:        1,
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
				RoomID:    1,
				StartDate: now,
				EndDate:   now.AddDate(0, 0, 2),
			},
			wantStatus: http.StatusOK,
		},
		{name: "missing session redirects to home", seed: nil, wantStatus: http.StatusSeeOther},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := newGET("/reservation-summary")
			if tc.seed != nil {
				session.Put(req.Context(), "reservation", *tc.seed)
			}
			rr := do(Repo.ReservationSummary, req)
			mustStatus(t, rr, tc.wantStatus)
		})
	}
}

// TestRepository_PostAvailability tests the room availability search functionality.
// This handler processes user date inputs, queries for available rooms, and either
// displays results or redirects with error messages. Tests cover date parsing,
// database errors, and both successful and unsuccessful searches.
func TestRepository_PostAvailability(t *testing.T) {
	t.Run("invalid start date", func(t *testing.T) {
		req := newPOSTForm("/search-availability", toForm(map[string]string{
			"start": "invalid",
			"end":   "01/02/2100",
		}))
		rr := do(Repo.PostAvailability, req)
		mustStatus(t, rr, http.StatusSeeOther)
		mustRedirectContains(t, rr, "/")
	})

	t.Run("invalid end date", func(t *testing.T) {
		req := newPOSTForm("/search-availability", toForm(map[string]string{
			"start": "01/01/2100",
			"end":   "invalid",
		}))
		rr := do(Repo.PostAvailability, req)
		mustStatus(t, rr, http.StatusSeeOther)
	})

	t.Run("database error during room search", func(t *testing.T) {
		dbrepo.ForceAllRoomsErr = true
		defer func() { dbrepo.ForceAllRoomsErr = false }()

		req := newPOSTForm("/search-availability", toForm(map[string]string{
			"start": "01/01/2100",
			"end":   "01/02/2100",
		}))
		rr := do(Repo.PostAvailability, req)
		mustStatus(t, rr, http.StatusSeeOther)
		mustRedirectContains(t, rr, "/")
	})

	t.Run("no rooms available for dates", func(t *testing.T) {
		req := newPOSTForm("/search-availability", toForm(map[string]string{
			"start": "01/01/2100", // test repo returns empty for these dates
			"end":   "01/02/2100",
		}))
		rr := do(Repo.PostAvailability, req)
		mustStatus(t, rr, http.StatusSeeOther)
		mustRedirectContains(t, rr, "/search-availability")
	})

	t.Run("rooms found for dates", func(t *testing.T) {
		req := newPOSTForm("/search-availability", toForm(map[string]string{
			"start": "01/01/2101", // test repo returns rooms for year 2101
			"end":   "01/02/2101",
		}))
		rr := do(Repo.PostAvailability, req)
		mustStatus(t, rr, http.StatusOK)
	})
}

// TestRepository_PostAvailability_ParseFormError tests malformed request body handling.
// This covers the case where the request body cannot be parsed as form data,
// which should result in a graceful error response.
func TestRepository_PostAvailability_ParseFormError(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/search-availability", strings.NewReader("%not-urlencoded"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = sessionize(req)
	rr := do(Repo.PostAvailability, req)
	mustStatus(t, rr, http.StatusSeeOther)
}

// TestRepository_AvailabilityJSON tests the AJAX availability checking endpoint.
// This endpoint returns JSON responses for real-time availability checking
// on individual room pages. Tests cover form parsing errors, database errors,
// and both available and unavailable scenarios.
func TestRepository_AvailabilityJSON(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantCode   int
		wantOK     *bool
		wantMsgSub string
	}{
		{"parse form error", "%not-urlencoded", http.StatusOK, ptrBool(false), "Internal server error"},
		{"database error (room 2)", "start=01/01/2102&end=01/02/2102&room_id=2", http.StatusOK, ptrBool(false), "Error querying database"},
		{"room not available", "start=01/01/2100&end=01/02/2100&room_id=1", http.StatusOK, ptrBool(false), ""},
		{"room available", "start=01/01/2101&end=01/02/2101&room_id=1", http.StatusOK, ptrBool(true), ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/search-availability-json", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req = sessionize(req)

			rr := do(Repo.AvailabilityJSON, req)
			mustStatus(t, rr, tc.wantCode)

			var resp jsonResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
				t.Fatalf("json unmarshal: %v", err)
			}
			if tc.wantOK != nil && resp.OK != *tc.wantOK {
				t.Fatalf("OK: got %v, want %v", resp.OK, *tc.wantOK)
			}
			if tc.wantMsgSub != "" && !strings.Contains(resp.Message, tc.wantMsgSub) {
				t.Fatalf("Message: got %q, want contains %q", resp.Message, tc.wantMsgSub)
			}
		})
	}
}

// TestRepository_ChooseRoom verifies room selection from availability results.
// This handler processes room selection after availability search, updating
// the session with the chosen room and redirecting to the reservation form.
// Tests cover URL parsing, session requirements, and invalid room IDs.
func TestRepository_ChooseRoom(t *testing.T) {
	tests := []struct {
		name       string
		roomID     string
		seedSess   bool
		wantStatus int
	}{
		{"valid room selection", "1", true, http.StatusSeeOther},
		{"invalid room id", "not-an-id", true, http.StatusSeeOther},
		{"missing session data", "1", false, http.StatusSeeOther},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := newGET("/choose-room/" + tc.roomID)
			// Handler parses RequestURI directly, so set it explicitly
			req.RequestURI = "/choose-room/" + tc.roomID

			if tc.seedSess {
				session.Put(req.Context(), "reservation", models.Reservation{RoomID: 1})
			}

			rr := do(Repo.ChooseRoom, req)
			mustStatus(t, rr, tc.wantStatus)
		})
	}
}

// TestRepository_BookRoom tests direct room booking from external links.
// This handler processes booking requests with room ID and dates in query parameters,
// typically used for direct booking links from room detail pages.
// Tests cover parameter parsing and room lookup validation.
func TestRepository_BookRoom(t *testing.T) {
	tests := []struct {
		name       string
		q          string
		wantStatus int
	}{
		{"valid booking request", "?id=1&s=01/01/2100&e=01/02/2100", http.StatusSeeOther},
		{"missing date parameters", "?id=1", http.StatusSeeOther},
		{"invalid room id", "?id=100&s=01/01/2100&e=01/02/2100", http.StatusSeeOther},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := newGET("/book-room" + tc.q)
			rr := do(Repo.BookRoom, req)
			mustStatus(t, rr, tc.wantStatus)
		})
	}
}

// TestRepository_ShowLogin verifies that the login page renders correctly.
// This is a simple test ensuring the login form is displayed without errors.
func TestRepository_ShowLogin(t *testing.T) {
	req := newGET("/user/login")
	rr := do(Repo.ShowLogin, req)
	mustStatus(t, rr, http.StatusOK)
}

// TestRepository_PostShowLogin_AuthFailure tests authentication failure handling.
// The test repo is configured to return an authentication error for the email
// "badlogin@example.com", which should result in a redirect back to the login page.
func TestRepository_PostShowLogin_AuthFailure(t *testing.T) {
	form := url.Values{}
	form.Set("email", "badlogin@example.com") // configured to fail in test repo
	form.Set("password", "doesntmatter")

	req := newPOSTForm("/user/login", form)
	rr := do(Repo.PostShowLogin, req)

	mustStatus(t, rr, http.StatusSeeOther)
	mustRedirectContains(t, rr, "/user/login")
}

// TestRepository_LoginRouteIntegration confirms the login route is properly wired in the router.
// This test verifies that the route configuration includes the login endpoint
// and that it's accessible without authentication.
func TestRepository_LoginRouteIntegration(t *testing.T) {
	ts := httptest.NewTLSServer(getRoutes())
	defer ts.Close()

	resp, err := ts.Client().Get(ts.URL + "/user/login")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
}

// TestRepository_PostShowLogin covers login form validation and successful authentication.
// Tests include missing required fields, invalid email format, and successful login
// scenarios. Successful authentication should redirect to the home page.
func TestRepository_PostShowLogin(t *testing.T) {
	tests := []struct {
		name       string
		email      string
		pass       string
		wantStatus int
	}{
		{"successful login", "test@example.com", "password", http.StatusSeeOther},
		{"missing email field", "", "password", http.StatusOK},      // re-renders form
		{"invalid email format", "bad@", "password", http.StatusOK}, // re-renders form
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			form := url.Values{}
			form.Set("email", tc.email)
			form.Set("password", tc.pass)
			req := newPOSTForm("/user/login", form)
			rr := do(Repo.PostShowLogin, req)
			mustStatus(t, rr, tc.wantStatus)
		})
	}
}

// TestRepository_Logout verifies session destruction and redirect behavior.
// The logout handler should destroy the current session and redirect to the login page.
func TestRepository_Logout(t *testing.T) {
	req := newGET("/user/logout")
	session.Put(req.Context(), "user_id", 1)
	rr := do(Repo.Logout, req)
	mustStatus(t, rr, http.StatusSeeOther)
}

// TestRepository_StaticRoomPages tests that static informational pages render correctly.
// These pages include room detail pages and general information pages that don't
// require complex data processing or user input.
func TestRepository_StaticRoomPages(t *testing.T) {
	pages := []struct {
		name string
		h    http.HandlerFunc
		u    string
	}{
		{"golden haybeam loft", Repo.GoldenHaybeamLoft, "/golden-haybeam-loft"},
		{"window perch theater", Repo.WindowPerchTheater, "/window-perch-theater"},
		{"laundry basket nook", Repo.LaundryBasketNook, "/laundry-basket-nook"},
		{"about", Repo.About, "/about"},
		{"photos", Repo.Photos, "/photos"},
		{"contact", Repo.Contact, "/contact"},
		{"home", Repo.Home, "/"},
	}
	for _, p := range pages {
		t.Run(p.name, func(t *testing.T) {
			req := newGET(p.u)
			rr := do(p.h, req)
			mustStatus(t, rr, http.StatusOK)
		})
	}
}

// TestRepository_AdminDashboard verifies the admin dashboard page renders correctly.
// This is the main administrative interface entry point.
func TestRepository_AdminDashboard(t *testing.T) {
	req := newGET("/admin/dashboard")
	rr := do(Repo.AdminDashboard, req)
	mustStatus(t, rr, http.StatusOK)
}

// TestRepository_AdminAllReservations verifies the complete reservations list displays correctly.
// This administrative page shows all reservations in the system for management purposes.
func TestRepository_AdminAllReservations(t *testing.T) {
	req := newGET("/admin/reservations-all")
	rr := do(Repo.AdminAllReservations, req)
	mustStatus(t, rr, http.StatusOK)
}

// TestRepository_AdminAllReservations_DBError tests database error handling in the reservations list.
// When the database query fails, the page should return a 500 error rather than crashing.
func TestRepository_AdminAllReservations_DBError(t *testing.T) {
	dbrepo.ForceAllReservationsErr = true
	defer func() { dbrepo.ForceAllReservationsErr = false }()

	req := newGET("/admin/reservations-all")
	rr := do(Repo.AdminAllReservations, req)
	mustStatus(t, rr, http.StatusInternalServerError)
}

// TestRepository_AdminNewReservations verifies the unprocessed reservations list displays correctly.
// This page shows reservations that require staff review and processing.
func TestRepository_AdminNewReservations(t *testing.T) {
	req := newGET("/admin/reservations-new")
	rr := do(Repo.AdminNewReservations, req)
	mustStatus(t, rr, http.StatusOK)
}

// TestRepository_AdminNewReservations_DBError tests database error handling in the new reservations list.
// When the database query fails, the page should return a 500 error rather than crashing.
func TestRepository_AdminNewReservations_DBError(t *testing.T) {
	dbrepo.ForceAllNewReservationsErr = true
	defer func() { dbrepo.ForceAllNewReservationsErr = false }()

	req := newGET("/admin/reservations-new")
	rr := do(Repo.AdminNewReservations, req)
	mustStatus(t, rr, http.StatusInternalServerError)
}

// TestRepository_AdminShowReservation verifies individual reservation detail page rendering.
// This page allows administrators to view and edit detailed reservation information.
// Tests cover valid reservations, invalid URLs, and reservations that don't exist.
func TestRepository_AdminShowReservation(t *testing.T) {
	tests := []struct {
		name       string
		reqURI     string
		q          string
		wantStatus int
	}{
		{"valid reservation", "/admin/reservations/new/1/show", "?y=2025&m=12", http.StatusOK},
		{"invalid reservation id", "/admin/reservations/new/invalid/show", "", http.StatusInternalServerError},
		{"reservation not found", "/admin/reservations/new/999/show", "", http.StatusOK},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := newGET(tc.reqURI + tc.q)
			// Handler parses RequestURI directly for path segments
			req.RequestURI = tc.reqURI
			rr := do(Repo.AdminShowReservation, req)
			mustStatus(t, rr, tc.wantStatus)
		})
	}
}

// TestRepository_AdminShowReservation_DBError tests database error handling in reservation details.
// When the reservation lookup fails, the page should return a 500 error.
func TestRepository_AdminShowReservation_DBError(t *testing.T) {
	dbrepo.ForceGetReservationErr = true
	defer func() { dbrepo.ForceGetReservationErr = false }()

	reqURI := "/admin/reservations/new/1/show"
	req := newGET(reqURI)
	req.RequestURI = reqURI
	rr := do(Repo.AdminShowReservation, req)
	mustStatus(t, rr, http.StatusInternalServerError)
}

// TestRepository_AdminPostShowReservation verifies reservation update form processing.
// This handler processes updates to reservation details from the administrative interface.
// Tests cover successful updates, invalid data, and different redirect destinations
// based on the source (list view vs calendar view).
func TestRepository_AdminPostShowReservation(t *testing.T) {
	tests := []struct {
		name       string
		reqURI     string
		form       map[string]string
		wantStatus int
	}{
		{
			name:   "successful update redirects to list",
			reqURI: "/admin/reservations/new/1/show",
			form: map[string]string{
				"first_name": "UpdatedJohn",
				"last_name":  "UpdatedDoe",
				"email":      "updated@example.com",
				"phone":      "1234567890",
				"month":      "", // empty means redirect to list
				"year":       "",
			},
			wantStatus: http.StatusSeeOther,
		},
		{
			name:   "successful update redirects to calendar",
			reqURI: "/admin/reservations/cal/1/show",
			form: map[string]string{
				"first_name": "CalendarJohn",
				"last_name":  "CalendarDoe",
				"email":      "calendar@example.com",
				"phone":      "1234567890",
				"month":      "12", // with month/year, redirects to calendar
				"year":       "2025",
			},
			wantStatus: http.StatusSeeOther,
		},
		{
			name:       "invalid reservation id",
			reqURI:     "/admin/reservations/new/invalid/show",
			form:       map[string]string{"first_name": "Test"},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "reservation not found still redirects",
			reqURI:     "/admin/reservations/new/999/show",
			form:       map[string]string{"first_name": "Test"},
			wantStatus: http.StatusSeeOther,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := newPOSTForm(tc.reqURI, toForm(tc.form))
			req.RequestURI = tc.reqURI
			rr := do(Repo.AdminPostShowReservation, req)
			mustStatus(t, rr, tc.wantStatus)
		})
	}
}

// TestRepository_AdminPostShowReservation_UpdateError tests database update error handling.
// When the reservation update fails in the database, the handler should return a 500 error.
func TestRepository_AdminPostShowReservation_UpdateError(t *testing.T) {
	dbrepo.ForceUpdateReservationErr = true
	defer func() { dbrepo.ForceUpdateReservationErr = false }()

	reqURI := "/admin/reservations/new/1/show"
	req := newPOSTForm(reqURI, toForm(map[string]string{
		"first_name": "X",
		"last_name":  "Y",
		"email":      "x@y.com",
		"phone":      "1",
	}))
	req.RequestURI = reqURI
	rr := do(Repo.AdminPostShowReservation, req)
	mustStatus(t, rr, http.StatusInternalServerError)
}

// TestRepository_AdminPostShowReservation_ParseFormError tests malformed form handling.
// When the request body cannot be parsed, the handler should return a 500 error.
func TestRepository_AdminPostShowReservation_ParseFormError(t *testing.T) {
	reqURI := "/admin/reservations/new/1/show"
	req := httptest.NewRequest(http.MethodPost, reqURI, strings.NewReader("%not-urlencoded"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = sessionize(req)
	req.RequestURI = reqURI
	rr := do(Repo.AdminPostShowReservation, req)
	mustStatus(t, rr, http.StatusInternalServerError)
}

// TestRepository_AdminReservationsCalendar_SessionSeeds ensures session data is properly stored.
// The calendar handler stores room block data in the session for later form processing.
// This test verifies that the session contains the expected data structure.
func TestRepository_AdminReservationsCalendar_SessionSeeds(t *testing.T) {
	req := newGET("/admin/reservations-calendar?y=2050&m=1")
	rr := do(Repo.AdminReservationsCalendar, req)
	mustStatus(t, rr, http.StatusOK)

	// Verify that block map data is stored in session for form processing
	val := session.Get(req.Context(), "block_map_1")
	m, ok := val.(map[string]int)
	if !ok || len(m) == 0 {
		t.Fatalf("expected non-empty block_map_1 in session; got %#v", val)
	}
}

// TestRepository_AdminReservationsCalendar verifies calendar page rendering.
// The calendar displays room availability with different views for current month
// and specific months specified via query parameters.
func TestRepository_AdminReservationsCalendar(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"current month display", "/admin/reservations-calendar"},
		{"specific month display", "/admin/reservations-calendar?y=2050&m=1"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := newGET(tc.url)
			rr := do(Repo.AdminReservationsCalendar, req)
			mustStatus(t, rr, http.StatusOK)
		})
	}
}

// TestRepository_AdminReservationsCalendar_AllRoomsError tests room data error handling.
// When the room lookup fails, the calendar page should return a 500 error.
func TestRepository_AdminReservationsCalendar_AllRoomsError(t *testing.T) {
	dbrepo.ForceAllRoomsErr = true
	defer func() { dbrepo.ForceAllRoomsErr = false }()

	req := newGET("/admin/reservations-calendar")
	rr := do(Repo.AdminReservationsCalendar, req)
	mustStatus(t, rr, http.StatusInternalServerError)
}

// TestRepository_AdminReservationsCalendar_RestrictionsError tests restrictions data error handling.
// When the room restrictions lookup fails, the calendar page should return a 500 error.
func TestRepository_AdminReservationsCalendar_RestrictionsError(t *testing.T) {
	dbrepo.ForceRestrictionsErr = true
	defer func() { dbrepo.ForceRestrictionsErr = false }()

	req := newGET("/admin/reservations-calendar?y=2050&m=1")
	rr := do(Repo.AdminReservationsCalendar, req)
	mustStatus(t, rr, http.StatusInternalServerError)
}

// TestRepository_AdminProcessReservation verifies the reservation processing workflow.
// This handler marks reservations as processed and redirects to the appropriate
// view (list or calendar) based on the source context and query parameters.
func TestRepository_AdminProcessReservation(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		id, src    string
		wantSubLoc string
	}{
		{"redirect to new reservations list", "/admin/process-reservation/new/1/do", "1", "new", "/admin/reservations-new"},
		{"redirect to calendar view", "/admin/process-reservation/new/1/do?y=2050&m=01", "1", "new", "/admin/reservations-calendar?y=2050&m=01"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := newGET(tc.url)
			// Set up chi route context with URL parameters
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.id)
			rctx.URLParams.Add("src", tc.src)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := do(Repo.AdminProcessReservation, req)
			mustStatus(t, rr, http.StatusSeeOther)
			mustRedirectContains(t, rr, tc.wantSubLoc)
		})
	}
}

// TestRepository_AdminProcessReservation_UpdateError tests processing error handling.
// When the database update fails, the handler should still redirect but log the error.
func TestRepository_AdminProcessReservation_UpdateError(t *testing.T) {
	dbrepo.ForceProcessedUpdateErr = true
	defer func() { dbrepo.ForceProcessedUpdateErr = false }()

	req := newGET("/admin/process-reservation/new/1/do")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	rctx.URLParams.Add("src", "new")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := do(Repo.AdminProcessReservation, req)
	// Handler logs error but still redirects for user experience
	mustStatus(t, rr, http.StatusSeeOther)
}

// TestRepository_AdminDeleteReservation verifies the reservation deletion workflow.
// This handler deletes reservations and redirects to the appropriate view
// based on the source context and query parameters.
func TestRepository_AdminDeleteReservation(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		id, src    string
		wantSubLoc string
	}{
		{"redirect to new reservations list", "/admin/delete-reservation/new/1/do", "1", "new", "/admin/reservations-new"},
		{"redirect to calendar view", "/admin/delete-reservation/new/1/do?y=2050&m=01", "1", "new", "/admin/reservations-calendar?y=2050&m=01"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := newGET(tc.url)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.id)
			rctx.URLParams.Add("src", tc.src)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := do(Repo.AdminDeleteReservation, req)
			mustStatus(t, rr, http.StatusSeeOther)
			mustRedirectContains(t, rr, tc.wantSubLoc)
		})
	}
}

// TestRepository_AdminPostReservationsCalendar tests calendar block management form processing.
// This handler processes calendar form submissions to add or remove room blocks.
// Tests cover basic saves, adding blocks, and removing blocks.
func TestRepository_AdminPostReservationsCalendar(t *testing.T) {
	tests := []struct {
		name       string
		form       url.Values
		wantSubLoc string
	}{
		{
			name:       "basic calendar save",
			form:       url.Values{"y": {"2050"}, "m": {"1"}},
			wantSubLoc: "/admin/reservations-calendar?y=2050&m=1",
		},
		{
			name: "add room block",
			form: url.Values{
				"y": {"2050"}, "m": {"1"},
				"add_block_1_01/01/2050": {""}, // checkbox for adding block
			},
			wantSubLoc: "/admin/reservations-calendar?y=2050&m=1",
		},
		{
			name: "remove room block",
			form: url.Values{
				"y": {"2050"}, "m": {"1"},
				"remove_block_1_01/01/2050": {""}, // checkbox for keeping existing block
			},
			wantSubLoc: "/admin/reservations-calendar?y=2050&m=1",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := newPOSTForm("/admin/reservations-calendar", tc.form)
			// Seed session with existing block data for processing
			session.Put(req.Context(), "block_map_1", map[string]int{"01/01/2050": 123})

			rr := do(Repo.AdminPostReservationsCalendar, req)
			mustStatus(t, rr, http.StatusSeeOther)
			mustRedirectContains(t, rr, tc.wantSubLoc)
		})
	}
}

// TestRepository_AdminPages_Router ensures admin routes are properly configured.
// This integration test verifies that administrative routes are accessible
// and return successful responses.
func TestRepository_AdminPages_Router(t *testing.T) {
	ts := httptest.NewTLSServer(getRoutes())
	defer ts.Close()

	paths := []string{
		"/admin/dashboard",
		"/admin/reservations-all",
		"/admin/reservations-new",
	}

	for _, p := range paths {
		t.Run(p, func(t *testing.T) {
			resp, err := ts.Client().Get(ts.URL + p)
			if err != nil {
				t.Fatalf("GET %s error: %v", p, err)
			}
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("GET %s: status %d want %d", p, resp.StatusCode, http.StatusOK)
			}
		})
	}
}

// TestRepository_AdminPostReservationsCalendar_ParseFormError tests form parsing error handling.
// When the calendar form cannot be parsed, the handler should return a 500 error.
func TestRepository_AdminPostReservationsCalendar_ParseFormError(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/admin/reservations-calendar", strings.NewReader("%not-urlencoded"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = sessionize(req)
	rr := do(Repo.AdminPostReservationsCalendar, req)
	mustStatus(t, rr, http.StatusInternalServerError)
}

// TestRepository_AdminPostReservationsCalendar_AllRoomsError tests room data error in calendar updates.
// When room lookup fails during calendar processing, the handler should return a 500 error.
func TestRepository_AdminPostReservationsCalendar_AllRoomsError(t *testing.T) {
	dbrepo.ForceAllRoomsErr = true
	defer func() { dbrepo.ForceAllRoomsErr = false }()

	form := url.Values{"y": {"2050"}, "m": {"1"}}
	req := newPOSTForm("/admin/reservations-calendar", form)
	rr := do(Repo.AdminPostReservationsCalendar, req)
	mustStatus(t, rr, http.StatusInternalServerError)
}

// TestRepository_AdminPostReservationsCalendar_DeleteBlockPath tests implicit block deletion.
// When blocks exist in session but are not marked for retention, they should be deleted.
func TestRepository_AdminPostReservationsCalendar_DeleteBlockPath(t *testing.T) {
	req := newPOSTForm("/admin/reservations-calendar", url.Values{"y": {"2050"}, "m": {"1"}})
	// Seed session with block that won't have a remove_block checkbox, triggering deletion
	session.Put(req.Context(), "block_map_1", map[string]int{"01/05/2050": 11})
	rr := do(Repo.AdminPostReservationsCalendar, req)
	mustStatus(t, rr, http.StatusSeeOther)
}

// TestRepository_AdminReservationsCalendar_WithReservationRestrictions tests reservation display in calendar.
// This test forces the test repo to include reservation restrictions, ensuring the calendar
// properly handles and displays both reservation blocks and owner blocks.
func TestRepository_AdminReservationsCalendar_WithReservationRestrictions(t *testing.T) {
	dbrepo.ForceHasReservationRestriction = true
	defer func() { dbrepo.ForceHasReservationRestriction = false }()

	req := newGET("/admin/reservations-calendar?y=2050&m=1")
	rr := do(Repo.AdminReservationsCalendar, req)
	mustStatus(t, rr, http.StatusOK)
}

// TestRepository_AdminPostShowReservation_GetReservationErr tests reservation lookup error during updates.
// When the reservation cannot be retrieved for updating, the handler should return a 500 error.
func TestRepository_AdminPostShowReservation_GetReservationErr(t *testing.T) {
	dbrepo.ForceGetReservationErr = true
	defer func() { dbrepo.ForceGetReservationErr = false }()

	reqURI := "/admin/reservations/new/1/show"
	req := newPOSTForm(reqURI, toForm(map[string]string{
		"first_name": "X",
		"last_name":  "Y",
		"email":      "x@y.com",
		"phone":      "1",
	}))
	req.RequestURI = reqURI
	rr := do(Repo.AdminPostShowReservation, req)
	mustStatus(t, rr, http.StatusInternalServerError)
}

// TestRepository_PostShowLogin_ParseFormError tests form parsing error in login processing.
// When the login form cannot be parsed, the handler should re-render the form
// rather than crashing.
func TestRepository_PostShowLogin_ParseFormError(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/user/login", strings.NewReader("%not-urlencoded"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = sessionize(req)

	rr := do(Repo.PostShowLogin, req)
	// Handler logs error but still renders form
	mustStatus(t, rr, http.StatusOK)
}

// TestRepository_PostReservation_InvalidForm_RoomLookupError tests room lookup failure during form re-rendering.
// When form validation fails and the subsequent room lookup for re-rendering also fails,
// the handler should redirect with an error rather than crashing.
func TestRepository_PostReservation_InvalidForm_RoomLookupError(t *testing.T) {
	req := newPOSTForm("/make-reservation", toForm(map[string]string{
		"start_date": "01/01/2100",
		"end_date":   "01/02/2100",
		"first_name": "J", // too short, causes validation failure
		"last_name":  "Smith",
		"email":      "john@smith.com",
		"phone":      "1234567891",
		"room_id":    "100", // triggers GetRoomByID error during form re-render
	}))
	rr := do(Repo.PostReservation, req)
	mustStatus(t, rr, http.StatusSeeOther)
}

// TestRepository_AdminPostReservationsCalendar_InsertBlockError tests block insertion error handling.
// When adding a new block fails in the database, the handler should log the error
// but continue processing and redirect normally.
func TestRepository_AdminPostReservationsCalendar_InsertBlockError(t *testing.T) {
	dbrepo.ForceInsertBlockErr = true
	defer func() { dbrepo.ForceInsertBlockErr = false }()

	form := url.Values{
		"y":                      {"2050"},
		"m":                      {"1"},
		"add_block_1_01/07/2050": {""}, // will trigger insert error
	}
	req := newPOSTForm("/admin/reservations-calendar", form)
	session.Put(req.Context(), "block_map_1", map[string]int{})

	rr := do(Repo.AdminPostReservationsCalendar, req)
	// Handler logs error but still redirects for user experience
	mustStatus(t, rr, http.StatusSeeOther)
}

// TestRepository_AdminPostReservationsCalendar_DeleteBlockError tests block deletion error handling.
// When removing a block fails in the database, the handler should log the error
// but continue processing and redirect normally.
func TestRepository_AdminPostReservationsCalendar_DeleteBlockError(t *testing.T) {
	dbrepo.ForceDeleteBlockErr = true
	defer func() { dbrepo.ForceDeleteBlockErr = false }()

	form := url.Values{"y": {"2050"}, "m": {"1"}}
	req := newPOSTForm("/admin/reservations-calendar", form)
	// Seed block that will be deleted (no remove checkbox), triggering delete error
	session.Put(req.Context(), "block_map_1", map[string]int{"01/05/2050": 11})

	rr := do(Repo.AdminPostReservationsCalendar, req)
	// Handler logs error but still redirects for user experience
	mustStatus(t, rr, http.StatusSeeOther)
}

// TestRepository_AdminShowReservation_ShortURL tests malformed URL handling.
// When the URL doesn't contain enough path segments, the handler should return
// a 500 error rather than panicking on array access.
func TestRepository_AdminShowReservation_ShortURL(t *testing.T) {
	reqURI := "/admin/reservations/new" // missing ID and action segments
	req := newGET(reqURI)
	req.RequestURI = reqURI
	rr := do(Repo.AdminShowReservation, req)
	mustStatus(t, rr, http.StatusInternalServerError)
}
