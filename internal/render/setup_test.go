package render

import (
	"encoding/gob"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/bensabler/milos-residence/internal/config"
	"github.com/bensabler/milos-residence/internal/models"
)

// session is the shared session manager used by tests in this package.
var session *scs.SessionManager

// testApp holds application configuration tailored for tests.
var testApp config.AppConfig

// TestMain initializes test-wide dependencies (logging, session, app config)
// and then executes the test suite.
func TestMain(m *testing.M) {

	// register types we plan to store in the session so gob can encode them
	// what I am going to put in the session
	gob.Register(models.Reservation{})

	// run in non-production mode for tests (e.g., no Secure cookies)
	// NOTE: Set to true for production to enable secure cookies and caching.
	testApp.InProduction = false

	// set up an INFO logger for operational test messages
	infoLog := log.New(os.Stdout, "INFO:\t", log.Ldate|log.Ltime)
	testApp.InfoLog = infoLog

	// set up an ERROR logger with file/line context for easier debugging
	errorLog := log.New(os.Stdout, "ERROR:\t", log.Ldate|log.Ltime|log.Lshortfile)
	testApp.ErrorLog = errorLog

	// --- Session configuration
	// create a new session manager and configure cookie behavior for tests
	session = scs.New()
	session.Lifetime = 24 * time.Hour // session lifetime during tests
	session.Cookie.Persist = true     // survive across requests like a browser
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = false // allow HTTP during tests (no TLS required)
	testApp.Session = session     // expose session via the test app config

	// make the render package use our test app configuration
	app = &testApp

	// run all tests and exit the process with the resulting status code
	os.Exit(m.Run())
}
