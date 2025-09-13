// Package render contains tests for renderer bootstrapping and shared test
// setup (session manager, config). Tests assume a non-production environment
// and configure an in-memory session for deterministic behavior.
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

// session is the test session manager configured in TestMain.
var session *scs.SessionManager

// testApp is the AppConfig used during tests. It is assigned to the package-
// level app pointer in TestMain so production code paths use it.
var testApp config.AppConfig

// TestMain sets up shared test fixtures: gob registrations, logging, and an
// scs session manager, then runs the test suite. It configures app to a
// non-production mode with SameSite Lax cookies and non-secure transport.
func TestMain(m *testing.M) {
	gob.Register(models.Reservation{})
	testApp.InProduction = false

	infoLog := log.New(os.Stdout, "INFO:\t", log.Ldate|log.Ltime)
	testApp.InfoLog = infoLog

	errorLog := log.New(os.Stdout, "ERROR:\t", log.Ldate|log.Ltime|log.Lshortfile)
	testApp.ErrorLog = errorLog

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = false
	testApp.Session = session

	// Point the package under test at our test configuration.
	app = &testApp

	os.Exit(m.Run())
}
