// Package helpers provides small, shared utilities for HTTP handlers and middleware.
// It centralizes consistent client/server error responses, global helper init,
// and an authentication check that relies on session state.
package helpers

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/bensabler/milos-residence/internal/config"
)

// app holds the process-wide application configuration used by helpers.
// Set once at startup via NewHelpers and treated as read-only thereafter.
var app *config.AppConfig

// NewHelpers wires the helpers package to the provided AppConfig.
// NewHelpers must be called during application initialization.
//
// Usage:
//
//	helpers.NewHelpers(&app)
func NewHelpers(a *config.AppConfig) {
	// Store the shared application config for use by helper functions.
	app = a
}

// ClientError writes a standardized client error response and logs the status.
// It produces an HTTP response with the supplied status code and default text.
//
// Parameters:
//   - w: response writer
//   - status: HTTP status code (e.g., http.StatusBadRequest)
//
// Side effects:
//   - Logs an informational line to app.InfoLog.
//   - Writes an HTTP error response to the client.
func ClientError(w http.ResponseWriter, status int) {
	// Record the client error with the chosen status code.
	app.InfoLog.Println("Client error with status of", status)

	// Emit a minimal error page with default status text.
	http.Error(w, http.StatusText(status), status)
}

// ServerError writes a standardized 500 response and logs a stack trace.
// It captures the current stack and the error message for diagnostics.
//
// Parameters:
//   - w: response writer
//   - err: triggering error
//
// Side effects:
//   - Logs a combined error + stack trace to app.ErrorLog.
//   - Writes a 500 Internal Server Error response to the client.
func ServerError(w http.ResponseWriter, err error) {
	// Compose error + stack trace to aid postmortem debugging.
	trace := fmt.Errorf("%s\n%s", err.Error(), debug.Stack())

	// Record the detailed trace in error logs.
	app.ErrorLog.Println(trace)

	// Return a generic 500 to the client without leaking internals.
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// IsAuthenticated reports whether the current request has an authenticated user.
// It checks for the presence of "user_id" in session state.
//
// Parameters:
//   - r: current HTTP request
//
// Returns:
//   - bool: true when a user_id exists in session; otherwise false.
func IsAuthenticated(r *http.Request) bool {
	// Lookup a user marker in the session to indicate authentication.
	exists := app.Session.Exists(r.Context(), "user_id")
	return exists
}
