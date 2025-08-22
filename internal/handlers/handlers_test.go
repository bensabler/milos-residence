package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
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
	params             []postData
	expectedStatusCode int
}{
	{"home", "/", "GET", []postData{}, http.StatusOK},
	{"about", "/about", "GET", []postData{}, http.StatusOK},
	{"photos", "/photos", "GET", []postData{}, http.StatusOK},
	{"golden-haybeam-loft", "/golden-haybeam-loft", "GET", []postData{}, http.StatusOK},
	{"window-perch-theater", "/window-perch-theater", "GET", []postData{}, http.StatusOK},
	{"laundry-basket-nook", "/laundry-basket-nook", "GET", []postData{}, http.StatusOK},
	{"search-availability", "/search-availability", "GET", []postData{}, http.StatusOK},
	{"contact", "/contact", "GET", []postData{}, http.StatusOK},
	{"make-reservation", "/make-reservation", "GET", []postData{}, http.StatusOK},
	{"reservation-summary", "/reservation-summary", "GET", []postData{}, http.StatusOK},
	{"post-search-availability", "/search-availability", "POST", []postData{
		{key: "start", value: "01-01-2020"},
		{key: "end", value: "01-02-2020"},
	}, http.StatusOK},
	{"post-search-availability-json", "/search-availability-json", "POST", []postData{
		{key: "start", value: "01-01-2020"},
		{key: "end", value: "01-02-2020"},
	}, http.StatusOK},
	{"make-reservation-post", "/make-reservation", "POST", []postData{
		{key: "first_name", value: "Ben"},
		{key: "last_name", value: "Sabler"},
		{key: "email", value: "ben@sabler.com"},
		{key: "phone", value: "999-888-7777"},
	}, http.StatusOK},
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
		if e.method == "GET" {
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
		} else {
			// build URL-encoded form data from the table of postData entries
			values := url.Values{}
			for _, x := range e.params {
				values.Add(x.key, x.value)
			}

			// send a POST with application/x-www-form-urlencoded body
			resp, err := ts.Client().PostForm(ts.URL+e.url, values)
			if err != nil {
				// log context for debugging and fail the test early
				t.Log(err)
				t.Fatal(err)
			}

			// verify the handler returned the expected status code
			if resp.StatusCode != e.expectedStatusCode {
				t.Errorf("for %s expected %d but got %d", e.name, e.expectedStatusCode, resp.StatusCode)
			}
		}
	}
}
