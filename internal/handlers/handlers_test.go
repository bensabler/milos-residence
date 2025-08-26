package handlers

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bensabler/milos-residence/internal/models"
)

// postData represents a single key/value form pair used in POST requests.
type postData struct {
	key   string
	value string
}

// theTests enumerates endpoints, methods, form params, and expected status codes.
// Each entry drives a single request against the test server to verify responses.
var theTests = []struct {
	name               string
	url                string
	method             string
	expectedStatusCode int
}{
	{"home", "/", "GET", http.StatusOK},
	{"about", "/about", "GET", http.StatusOK},
	{"photos", "/photos", "GET", http.StatusOK},
	{"golden-haybeam-loft", "/golden-haybeam-loft", "GET", http.StatusOK},
	{"window-perch-theater", "/window-perch-theater", "GET", http.StatusOK},
	{"laundry-basket-nook", "/laundry-basket-nook", "GET", http.StatusOK},
	{"search-availability", "/search-availability", "GET", http.StatusOK},
	{"contact", "/contact", "GET", http.StatusOK},
	{"make-reservation", "/make-reservation", "GET", http.StatusOK},
	{"reservation-summary", "/reservation-summary", "GET", http.StatusOK},
	// {"post-search-availability", "/search-availability", "POST", []postData{
	// 	{key: "start", value: "01-01-2020"},
	// 	{key: "end", value: "01-02-2020"},
	// }, http.StatusOK},
	// {"post-search-availability-json", "/search-availability-json", "POST", []postData{
	// 	{key: "start", value: "01-01-2020"},
	// 	{key: "end", value: "01-02-2020"},
	// }, http.StatusOK},
	// {"make-reservation-post", "/make-reservation", "POST", []postData{
	// 	{key: "first_name", value: "Ben"},
	// 	{key: "last_name", value: "Sabler"},
	// 	{key: "email", value: "ben@sabler.com"},
	// 	{key: "phone", value: "999-888-7777"},
	// }, http.StatusOK},
}

// TestHandlers spins up a TLS test server and exercises each route/method pair.
// It sends either GET or POST requests based on the table above and asserts the status code.
func TestHandlers(t *testing.T) {
	// build the app routes and start an HTTPS test server
	routes := getRoutes()
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	// iterate over each test case and perform the appropriate request
	for _, e := range theTests {
		// send a GET request to the target path
		resp, err := ts.Client().Get(ts.URL + e.url)
		if err != nil {
			// log context for debugging and fail the test early
			t.Log(err)
			t.Fatal(err)
		}

		// verify we received the expected HTTP status
		if resp.StatusCode != e.expectedStatusCode {
			t.Errorf("for %s expected %d but got %d", e.name, e.expectedStatusCode, resp.StatusCode)
		}

		// verify the handler returned the expected status code
		if resp.StatusCode != e.expectedStatusCode {
			t.Errorf("for %s expected %d but got %d", e.name, e.expectedStatusCode, resp.StatusCode)
		}
	}
}

func TestRepository_Reservation(t *testing.T) {
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "Golden Haybeam Loft",
		},
	}

	req, _ := http.NewRequest("GET", "/make-reservation", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)

	handler := http.HandlerFunc(Repo.MakeReservation)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Reserveation handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusOK)
	}

	// test case where reservation is not in session (reset everything)
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reserveation handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

	// test with non-existent room
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()
	reservation.RoomID = 100
	session.Put(ctx, "reservation", reservation)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reserveation handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

}

func getCtx(req *http.Request) context.Context {
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		log.Println(err)
	}
	return ctx
}
