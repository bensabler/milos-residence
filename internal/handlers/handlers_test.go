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

//
// -----------------------------------------------------------------------------
// Small helpers (DRY) — tiny, focused, and documented
// -----------------------------------------------------------------------------

// sessionize attaches the session context to a request for use in handler tests.
func sessionize(req *http.Request) *http.Request {
	ctx, _ := session.Load(req.Context(), req.Header.Get("X-Session"))
	return req.WithContext(ctx)
}

// newGET builds a GET request with session attached.
func newGET(path string) *http.Request {
	return sessionize(httptest.NewRequest(http.MethodGet, path, nil))
}

// newPOSTForm builds a POST request with form data and session attached.
func newPOSTForm(path string, form url.Values) *http.Request {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return sessionize(req)
}

// do executes a handler and returns a recorder with the response.
func do(h http.HandlerFunc, req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

// mustStatus fails the test if status code mismatches.
func mustStatus(t *testing.T, rr *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rr.Code != want {
		t.Fatalf("status: got %d, want %d", rr.Code, want)
	}
}

// mustRedirectContains asserts the Location header includes a substring.
func mustRedirectContains(t *testing.T, rr *httptest.ResponseRecorder, wantSub string) {
	t.Helper()
	loc := rr.Header().Get("Location")
	if !strings.Contains(loc, wantSub) {
		t.Fatalf("redirect: got %q, want contains %q", loc, wantSub)
	}
}

// toForm converts a map into url.Values.
func toForm(m map[string]string) url.Values {
	v := url.Values{}
	for k, val := range m {
		v.Set(k, val)
	}
	return v
}

// ptrBool returns a pointer to a bool literal (for table tests).
func ptrBool(b bool) *bool { return &b }

//
// -----------------------------------------------------------------------------
// Constructors — smoke tests
// -----------------------------------------------------------------------------

// TestNewRepo verifies NewRepo constructs a Repository with a non-nil DB and the provided AppConfig.
func TestNewRepo(t *testing.T) {
	app := &config.AppConfig{}
	// NewRepo just wires the repo; a zero sql.DB is fine for the unit test.
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

// TestNewHandlers verifies NewHandlers sets the global Repo pointer.
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

//
// -----------------------------------------------------------------------------
// Router — smoke test for public routes
// -----------------------------------------------------------------------------

// TestRoutes_Smoke ensures public routes are registered and return HTTP 200 OK.
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

//
// -----------------------------------------------------------------------------
// Reservation flow
// -----------------------------------------------------------------------------

// TestRepository_MakeReservation verifies session preconditions and room lookup.
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
				RoomID: 100, // test repo returns error
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

// TestRepository_PostReservation_ParseFormAndRoomLookupErrors exercises ParseForm failure and room lookup failure after validation succeeds.
func TestRepository_PostReservation_ParseFormAndRoomLookupErrors(t *testing.T) {
	// Malformed form -> ParseForm error path
	req := httptest.NewRequest(http.MethodPost, "/make-reservation", strings.NewReader("%not-urlencoded"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = sessionize(req)
	rr := do(Repo.PostReservation, req)
	mustStatus(t, rr, http.StatusSeeOther)

	// Valid form but room lookup fails after validation (room_id > 3)
	req = newPOSTForm("/make-reservation", toForm(map[string]string{
		"start_date": "01/01/2100",
		"end_date":   "01/02/2100",
		"first_name": "John",
		"last_name":  "Smith",
		"email":      "john@smith.com",
		"phone":      "1234567891",
		"room_id":    "100", // triggers GetRoomByID error post-validation
	}))
	rr = do(Repo.PostReservation, req)
	mustStatus(t, rr, http.StatusSeeOther)
}

// TestRepository_PostReservation covers validation, DB errors, and success paths.
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
			name: "validation failure (short first name)",
			form: map[string]string{
				"start_date": "01/01/2100",
				"end_date":   "01/02/2100",
				"first_name": "J",
				"last_name":  "Smith",
				"email":      "john@smith.com",
				"phone":      "1234567891",
				"room_id":    "1",
			},
			wantStatus: http.StatusOK, // re-render form
		},
		{
			name: "insert reservation DB error",
			form: map[string]string{
				"start_date": "01/01/2100",
				"end_date":   "01/02/2100",
				"first_name": "John",
				"last_name":  "Smith",
				"email":      "john@smith.com",
				"phone":      "1234567891",
				"room_id":    "2", // triggers test repo error
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
				"room_id":    "3",
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
			name: "invalid room_id (non-integer)",
			form: map[string]string{
				"start_date": "01/01/2100",
				"end_date":   "01/02/2100",
				"first_name": "John",
				"last_name":  "Smith",
				"email":      "john@smith.com",
				"phone":      "1234567891",
				"room_id":    "x",
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

// TestRepository_ReservationSummary verifies summary rendering and missing session handling.
func TestRepository_ReservationSummary(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name       string
		seed       *models.Reservation
		wantStatus int
	}{
		{
			name: "ok",
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
		{name: "missing session", seed: nil, wantStatus: http.StatusSeeOther},
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

//
// -----------------------------------------------------------------------------
// Availability endpoints
// -----------------------------------------------------------------------------

// TestRepository_PostAvailability exercises invalid date, empty results, rooms found, and DB error toggle.
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

	t.Run("DB error", func(t *testing.T) {
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

	t.Run("no rooms", func(t *testing.T) {
		req := newPOSTForm("/search-availability", toForm(map[string]string{
			"start": "01/01/2100",
			"end":   "01/02/2100",
		}))
		rr := do(Repo.PostAvailability, req)
		mustStatus(t, rr, http.StatusSeeOther)
		mustRedirectContains(t, rr, "/search-availability")
	})

	t.Run("rooms found", func(t *testing.T) {
		req := newPOSTForm("/search-availability", toForm(map[string]string{
			"start": "01/01/2101",
			"end":   "01/02/2101",
		}))
		rr := do(Repo.PostAvailability, req)
		mustStatus(t, rr, http.StatusOK)
	})
}

// TestRepository_PostAvailability_ParseFormError covers malformed body parsing.
func TestRepository_PostAvailability_ParseFormError(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/search-availability", strings.NewReader("%not-urlencoded"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = sessionize(req)
	rr := do(Repo.PostAvailability, req)
	mustStatus(t, rr, http.StatusSeeOther)
}

// TestRepository_AvailabilityJSON covers parse error, DB error, and ok/no paths.
func TestRepository_AvailabilityJSON(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantCode   int
		wantOK     *bool
		wantMsgSub string
	}{
		{"parse form error", "%not-urlencoded", http.StatusOK, ptrBool(false), "Internal server error"},
		{"DB error (room 2)", "start=01/01/2102&end=01/02/2102&room_id=2", http.StatusOK, ptrBool(false), "Error querying database"},
		{"OK=false", "start=01/01/2100&end=01/02/2100&room_id=1", http.StatusOK, ptrBool(false), ""},
		{"OK=true", "start=01/01/2101&end=01/02/2101&room_id=1", http.StatusOK, ptrBool(true), ""},
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

//
// -----------------------------------------------------------------------------
// Choose / Book
// -----------------------------------------------------------------------------

// TestRepository_ChooseRoom verifies URL parsing and session updates.
func TestRepository_ChooseRoom(t *testing.T) {
	tests := []struct {
		name       string
		roomID     string
		seedSess   bool
		wantStatus int
	}{
		{"valid", "1", true, http.StatusSeeOther},
		{"invalid id", "not-an-id", true, http.StatusSeeOther},
		{"missing session", "1", false, http.StatusSeeOther},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := newGET("/choose-room/" + tc.roomID)
			// Handler parses RequestURI directly.
			req.RequestURI = "/choose-room/" + tc.roomID

			if tc.seedSess {
				session.Put(req.Context(), "reservation", models.Reservation{RoomID: 1})
			}

			rr := do(Repo.ChooseRoom, req)
			mustStatus(t, rr, tc.wantStatus)
		})
	}
}

// TestRepository_BookRoom checks query parsing and reservation setup.
func TestRepository_BookRoom(t *testing.T) {
	tests := []struct {
		name       string
		q          string
		wantStatus int
	}{
		{"valid", "?id=1&s=01/01/2100&e=01/02/2100", http.StatusSeeOther},
		{"missing params", "?id=1", http.StatusSeeOther},
		{"room lookup error", "?id=100&s=01/01/2100&e=01/02/2100", http.StatusSeeOther},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := newGET("/book-room" + tc.q)
			rr := do(Repo.BookRoom, req)
			mustStatus(t, rr, tc.wantStatus)
		})
	}
}

//
// -----------------------------------------------------------------------------
// Auth
// -----------------------------------------------------------------------------

// TestRepository_ShowLogin verifies login page renders.
func TestRepository_ShowLogin(t *testing.T) {
	req := newGET("/user/login")
	rr := do(Repo.ShowLogin, req)
	mustStatus(t, rr, http.StatusOK)
}

// TestRepository_PostShowLogin_AuthFailure toggles the test repo for auth failure and asserts redirect.
func TestRepository_PostShowLogin_AuthFailure(t *testing.T) {
	form := url.Values{}
	form.Set("email", "badlogin@example.com") // forces Authenticate error in test repo
	form.Set("password", "doesntmatter")

	req := newPOSTForm("/user/login", form)
	rr := do(Repo.PostShowLogin, req)

	mustStatus(t, rr, http.StatusSeeOther)
	mustRedirectContains(t, rr, "/user/login")
}

// TestRepository_LoginRouteIntegration confirms the /user/login route is wired in the router.
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

// TestRepository_PostShowLogin covers validation and successful login.
func TestRepository_PostShowLogin(t *testing.T) {
	tests := []struct {
		name       string
		email      string
		pass       string
		wantStatus int
	}{
		{"success", "test@example.com", "password", http.StatusSeeOther},
		{"missing email", "", "password", http.StatusOK},
		{"invalid email", "bad@", "password", http.StatusOK},
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

// TestRepository_Logout verifies session destroy + redirect.
func TestRepository_Logout(t *testing.T) {
	req := newGET("/user/logout")
	session.Put(req.Context(), "user_id", 1)
	rr := do(Repo.Logout, req)
	mustStatus(t, rr, http.StatusSeeOther)
}

//
// -----------------------------------------------------------------------------
// Admin pages
// -----------------------------------------------------------------------------

// TestRepository_StaticRoomPages renders room and info pages directly.
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

// TestRepository_AdminDashboard verifies the admin dashboard page.
func TestRepository_AdminDashboard(t *testing.T) {
	req := newGET("/admin/dashboard")
	rr := do(Repo.AdminDashboard, req)
	mustStatus(t, rr, http.StatusOK)
}

// TestRepository_AdminAllReservations verifies list of all reservations.
func TestRepository_AdminAllReservations(t *testing.T) {
	req := newGET("/admin/reservations-all")
	rr := do(Repo.AdminAllReservations, req)
	mustStatus(t, rr, http.StatusOK)
}

// TestRepository_AdminAllReservations_DBError uses the toggle to force an error.
func TestRepository_AdminAllReservations_DBError(t *testing.T) {
	dbrepo.ForceAllReservationsErr = true
	defer func() { dbrepo.ForceAllReservationsErr = false }()

	req := newGET("/admin/reservations-all")
	rr := do(Repo.AdminAllReservations, req)
	mustStatus(t, rr, http.StatusInternalServerError)
}

// TestRepository_AdminNewReservations verifies list of new reservations.
func TestRepository_AdminNewReservations(t *testing.T) {
	req := newGET("/admin/reservations-new")
	rr := do(Repo.AdminNewReservations, req)
	mustStatus(t, rr, http.StatusOK)
}

// TestRepository_AdminNewReservations_DBError uses the toggle to force an error.
func TestRepository_AdminNewReservations_DBError(t *testing.T) {
	dbrepo.ForceAllNewReservationsErr = true
	defer func() { dbrepo.ForceAllNewReservationsErr = false }()

	req := newGET("/admin/reservations-new")
	rr := do(Repo.AdminNewReservations, req)
	mustStatus(t, rr, http.StatusInternalServerError)
}

// TestRepository_AdminShowReservation verifies admin detail view parsing and render.
func TestRepository_AdminShowReservation(t *testing.T) {
	tests := []struct {
		name       string
		reqURI     string
		q          string
		wantStatus int
	}{
		{"valid", "/admin/reservations/new/1/show", "?y=2025&m=12", http.StatusOK},
		{"invalid id", "/admin/reservations/new/invalid/show", "", http.StatusInternalServerError},
		{"not found but ok", "/admin/reservations/new/999/show", "", http.StatusOK},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := newGET(tc.reqURI + tc.q)
			// Handler parses RequestURI; set it exactly.
			req.RequestURI = tc.reqURI
			rr := do(Repo.AdminShowReservation, req)
			mustStatus(t, rr, tc.wantStatus)
		})
	}
}

// TestRepository_AdminShowReservation_DBError forces repo error.
func TestRepository_AdminShowReservation_DBError(t *testing.T) {
	dbrepo.ForceGetReservationErr = true
	defer func() { dbrepo.ForceGetReservationErr = false }()

	reqURI := "/admin/reservations/new/1/show"
	req := newGET(reqURI)
	req.RequestURI = reqURI
	rr := do(Repo.AdminShowReservation, req)
	mustStatus(t, rr, http.StatusInternalServerError)
}

// TestRepository_AdminPostShowReservation verifies update and redirect logic.
func TestRepository_AdminPostShowReservation(t *testing.T) {
	tests := []struct {
		name       string
		reqURI     string
		form       map[string]string
		wantStatus int
	}{
		{
			name:   "update ok -> list",
			reqURI: "/admin/reservations/new/1/show",
			form: map[string]string{
				"first_name": "UpdatedJohn",
				"last_name":  "UpdatedDoe",
				"email":      "updated@example.com",
				"phone":      "1234567890",
				"month":      "",
				"year":       "",
			},
			wantStatus: http.StatusSeeOther,
		},
		{
			name:   "update ok -> calendar",
			reqURI: "/admin/reservations/cal/1/show",
			form: map[string]string{
				"first_name": "CalendarJohn",
				"last_name":  "CalendarDoe",
				"email":      "calendar@example.com",
				"phone":      "1234567890",
				"month":      "12",
				"year":       "2025",
			},
			wantStatus: http.StatusSeeOther,
		},
		{
			name:       "invalid id",
			reqURI:     "/admin/reservations/new/invalid/show",
			form:       map[string]string{"first_name": "Test"},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "not found -> still redirects (test repo returns empty)",
			reqURI:     "/admin/reservations/new/999/show",
			form:       map[string]string{"first_name": "Test"},
			wantStatus: http.StatusSeeOther,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := newPOSTForm(tc.reqURI, toForm(tc.form))
			req.RequestURI = tc.reqURI // handler parses RequestURI directly
			rr := do(Repo.AdminPostShowReservation, req)
			mustStatus(t, rr, tc.wantStatus)
		})
	}
}

// TestRepository_AdminPostShowReservation_UpdateError toggles DB update error path.
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

// TestRepository_AdminPostShowReservation_ParseFormError covers malformed body parsing.
func TestRepository_AdminPostShowReservation_ParseFormError(t *testing.T) {
	reqURI := "/admin/reservations/new/1/show"
	req := httptest.NewRequest(http.MethodPost, reqURI, strings.NewReader("%not-urlencoded"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = sessionize(req)
	req.RequestURI = reqURI
	rr := do(Repo.AdminPostShowReservation, req)
	mustStatus(t, rr, http.StatusInternalServerError)
}

// TestRepository_AdminReservationsCalendar_SessionSeeds ensures per-room maps are saved to session.
func TestRepository_AdminReservationsCalendar_SessionSeeds(t *testing.T) {
	req := newGET("/admin/reservations-calendar?y=2050&m=1")
	rr := do(Repo.AdminReservationsCalendar, req)
	mustStatus(t, rr, http.StatusOK)

	// AllRooms returns room 1 in test repo; block_map_1 should be populated.
	val := session.Get(req.Context(), "block_map_1")
	m, ok := val.(map[string]int)
	if !ok || len(m) == 0 {
		t.Fatalf("expected non-empty block_map_1 in session; got %#v", val)
	}
}

// TestRepository_AdminReservationsCalendar verifies default and custom month renders.
func TestRepository_AdminReservationsCalendar(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"default month", "/admin/reservations-calendar"},
		{"custom month", "/admin/reservations-calendar?y=2050&m=1"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := newGET(tc.url)
			rr := do(Repo.AdminReservationsCalendar, req)
			mustStatus(t, rr, http.StatusOK)
		})
	}
}

// TestRepository_AdminReservationsCalendar_AllRoomsError toggles an error from AllRooms.
func TestRepository_AdminReservationsCalendar_AllRoomsError(t *testing.T) {
	dbrepo.ForceAllRoomsErr = true
	defer func() { dbrepo.ForceAllRoomsErr = false }()

	req := newGET("/admin/reservations-calendar")
	rr := do(Repo.AdminReservationsCalendar, req)
	mustStatus(t, rr, http.StatusInternalServerError)
}

// TestRepository_AdminReservationsCalendar_RestrictionsError toggles an error from GetRestrictionsForRoomByDate.
func TestRepository_AdminReservationsCalendar_RestrictionsError(t *testing.T) {
	dbrepo.ForceRestrictionsErr = true
	defer func() { dbrepo.ForceRestrictionsErr = false }()

	req := newGET("/admin/reservations-calendar?y=2050&m=1")
	rr := do(Repo.AdminReservationsCalendar, req)
	mustStatus(t, rr, http.StatusInternalServerError)
}

// TestRepository_AdminProcessReservation verifies redirect behavior (list vs calendar).
func TestRepository_AdminProcessReservation(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		id, src    string
		wantSubLoc string
	}{
		{"to new list", "/admin/process-reservation/new/1/do", "1", "new", "/admin/reservations-new"},
		{"to calendar", "/admin/process-reservation/new/1/do?y=2050&m=01", "1", "new", "/admin/reservations-calendar?y=2050&m=01"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := newGET(tc.url)
			// Set chi params.
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

// TestRepository_AdminProcessReservation_UpdateError toggles error during "mark processed".
func TestRepository_AdminProcessReservation_UpdateError(t *testing.T) {
	dbrepo.ForceProcessedUpdateErr = true
	defer func() { dbrepo.ForceProcessedUpdateErr = false }()

	req := newGET("/admin/process-reservation/new/1/do")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	rctx.URLParams.Add("src", "new")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := do(Repo.AdminProcessReservation, req)
	// Handler logs error but still redirects
	mustStatus(t, rr, http.StatusSeeOther)
}

// TestRepository_AdminDeleteReservation verifies deletion redirects.
func TestRepository_AdminDeleteReservation(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		id, src    string
		wantSubLoc string
	}{
		{"to new list", "/admin/delete-reservation/new/1/do", "1", "new", "/admin/reservations-new"},
		{"to calendar", "/admin/delete-reservation/new/1/do?y=2050&m=01", "1", "new", "/admin/reservations-calendar?y=2050&m=01"},
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

// TestRepository_AdminPostReservationsCalendar exercises add/remove block loops and redirect.
func TestRepository_AdminPostReservationsCalendar(t *testing.T) {
	tests := []struct {
		name       string
		form       url.Values
		wantSubLoc string
	}{
		{
			name:       "basic save",
			form:       url.Values{"y": {"2050"}, "m": {"1"}},
			wantSubLoc: "/admin/reservations-calendar?y=2050&m=1",
		},
		{
			name: "add block",
			form: url.Values{
				"y": {"2050"}, "m": {"1"},
				"add_block_1_01/01/2050": {""},
			},
			wantSubLoc: "/admin/reservations-calendar?y=2050&m=1",
		},
		{
			name: "remove block (kept because flag present)",
			form: url.Values{
				"y": {"2050"}, "m": {"1"},
				"remove_block_1_01/01/2050": {""},
			},
			wantSubLoc: "/admin/reservations-calendar?y=2050&m=1",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := newPOSTForm("/admin/reservations-calendar", tc.form)
			// Seed a block so the "remove" scan iterates an existing entry.
			session.Put(req.Context(), "block_map_1", map[string]int{"01/01/2050": 123})

			rr := do(Repo.AdminPostReservationsCalendar, req)
			mustStatus(t, rr, http.StatusSeeOther)
			mustRedirectContains(t, rr, tc.wantSubLoc)
		})
	}
}

// TestRepository_AdminPages_Router ensures top-level admin routes are wired.
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

// TestRepository_AdminPostReservationsCalendar_ParseFormError covers malformed body parsing.
func TestRepository_AdminPostReservationsCalendar_ParseFormError(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/admin/reservations-calendar", strings.NewReader("%not-urlencoded"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = sessionize(req)
	rr := do(Repo.AdminPostReservationsCalendar, req)
	mustStatus(t, rr, http.StatusInternalServerError)
}

// TestRepository_AdminPostReservationsCalendar_AllRoomsError toggles error returned by AllRooms.
func TestRepository_AdminPostReservationsCalendar_AllRoomsError(t *testing.T) {
	dbrepo.ForceAllRoomsErr = true
	defer func() { dbrepo.ForceAllRoomsErr = false }()

	form := url.Values{"y": {"2050"}, "m": {"1"}}
	req := newPOSTForm("/admin/reservations-calendar", form)
	rr := do(Repo.AdminPostReservationsCalendar, req)
	mustStatus(t, rr, http.StatusInternalServerError)
}

// TestRepository_AdminPostReservationsCalendar_DeleteBlockPath ensures the implicit delete branch is reached.
func TestRepository_AdminPostReservationsCalendar_DeleteBlockPath(t *testing.T) {
	// Seed a block in session; not sending a remove_* flag triggers DeleteBlockByID.
	req := newPOSTForm("/admin/reservations-calendar", url.Values{"y": {"2050"}, "m": {"1"}})
	session.Put(req.Context(), "block_map_1", map[string]int{"01/05/2050": 11})
	rr := do(Repo.AdminPostReservationsCalendar, req)
	mustStatus(t, rr, http.StatusSeeOther)
}

// Add these to internal/handlers/handlers_test.go

// ...imports and existing tests remain...

// Ensures the calendar code path that marks reservationMap (ReservationID > 0) is executed.
func TestRepository_AdminReservationsCalendar_WithReservationRestrictions(t *testing.T) {
	dbrepo.ForceHasReservationRestriction = true
	defer func() { dbrepo.ForceHasReservationRestriction = false }()

	req := newGET("/admin/reservations-calendar?y=2050&m=1")
	rr := do(Repo.AdminReservationsCalendar, req)
	mustStatus(t, rr, http.StatusOK)
}

// Forces GetReservationByID to fail in AdminPostShowReservation.
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

// Hits the ParseForm error branch in PostShowLogin (handler logs and re-renders with 200).
func TestRepository_PostShowLogin_ParseFormError(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/user/login", strings.NewReader("%not-urlencoded"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = sessionize(req)

	rr := do(Repo.PostShowLogin, req)
	mustStatus(t, rr, http.StatusOK)
}

// Covers the "form invalid + room lookup fails" branch in PostReservation.
// Handler attempts DB.GetRoomByID to re-render the form, gets an error, and redirects.
func TestRepository_PostReservation_InvalidForm_RoomLookupError(t *testing.T) {
	req := newPOSTForm("/make-reservation", toForm(map[string]string{
		"start_date": "01/01/2100",
		"end_date":   "01/02/2100",
		"first_name": "J", // too short -> invalid form
		"last_name":  "Smith",
		"email":      "john@smith.com",
		"phone":      "1234567891",
		"room_id":    "100", // triggers GetRoomByID error in invalid-form path
	}))
	rr := do(Repo.PostReservation, req)
	mustStatus(t, rr, http.StatusSeeOther)
}

// Forces the calendar add-block error log path (InsertBlockForRoom returns error).
func TestRepository_AdminPostReservationsCalendar_InsertBlockError(t *testing.T) {
	dbrepo.ForceInsertBlockErr = true
	defer func() { dbrepo.ForceInsertBlockErr = false }()

	form := url.Values{
		"y": {"2050"},
		"m": {"1"},
		// Add a block so handler attempts InsertBlockForRoom and logs the error.
		"add_block_1_01/07/2050": {""},
	}
	req := newPOSTForm("/admin/reservations-calendar", form)
	// Seed any block map; not strictly required for add path.
	session.Put(req.Context(), "block_map_1", map[string]int{})

	rr := do(Repo.AdminPostReservationsCalendar, req)
	mustStatus(t, rr, http.StatusSeeOther)
}

// Forces the calendar delete-block error log path (DeleteBlockByID returns error).
func TestRepository_AdminPostReservationsCalendar_DeleteBlockError(t *testing.T) {
	dbrepo.ForceDeleteBlockErr = true
	defer func() { dbrepo.ForceDeleteBlockErr = false }()

	// Post WITHOUT the remove flag so handler tries DeleteBlockByID and logs error.
	form := url.Values{"y": {"2050"}, "m": {"1"}}
	req := newPOSTForm("/admin/reservations-calendar", form)
	session.Put(req.Context(), "block_map_1", map[string]int{"01/05/2050": 11})

	rr := do(Repo.AdminPostReservationsCalendar, req)
	mustStatus(t, rr, http.StatusSeeOther)
}

// Covers the "Not enough URL parts" branch in AdminShowReservation.
func TestRepository_AdminShowReservation_ShortURL(t *testing.T) {
	reqURI := "/admin/reservations/new" // too few segments; len(exploded) <= 4
	req := newGET(reqURI)
	req.RequestURI = reqURI // handler reads RequestURI directly
	rr := do(Repo.AdminShowReservation, req)
	mustStatus(t, rr, http.StatusInternalServerError)
}

// When validation fails AND GetRoomByID (used to re-render the form) errors,
// the handler should redirect with SeeOther instead of rendering.
func TestRepository_PostReservation_InvalidForm_GetRoomError(t *testing.T) {
	req := newPOSTForm("/make-reservation", toForm(map[string]string{
		"start_date": "01/01/2100",
		"end_date":   "01/02/2100",
		"first_name": "J", // too short -> invalid form
		"last_name":  "Smith",
		"email":      "john@smith.com",
		"phone":      "1234567891",
		"room_id":    "100", // triggers GetRoomByID error in the invalid-form path
	}))
	rr := do(Repo.PostReservation, req)
	mustStatus(t, rr, http.StatusSeeOther)
}
