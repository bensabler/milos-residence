package helpers

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/bensabler/milos-residence/internal/config"
)

var app *config.AppConfig

// NewHelpers connects the helpers package to the shared application configuration.
func NewHelpers(a *config.AppConfig) {
	// keep a pointer to the app config so helper functions can use logs, session, etc.
	app = a
}

// ClientError logs a client error and writes the corresponding HTTP status to the response.
func ClientError(w http.ResponseWriter, status int) {
	// make a note in the info log so operators can see what the client experienced
	app.InfoLog.Println("Client error with status of", status)

	// send the standard status text (e.g., "Bad Request") with the given code
	http.Error(w, http.StatusText(status), status)

	// TODO: consider including a request ID (if you add one to context) to aid tracing.
}

// ServerError logs the error and stack trace, then returns a 500 response.
func ServerError(w http.ResponseWriter, err error) {
	// combine the original error with a snapshot of the current call stack
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())

	// record the detailed trace to the centralized error logger for diagnosis
	app.ErrorLog.Println(trace)

	// reply to the client with a generic 500 so we don’t expose internals
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

	// TODO: ensure verbose traces are gated by environment and never shown to users.
}
