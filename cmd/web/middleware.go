// Command web defines HTTP middleware used by the application binary.
// It provides CSRF protection (NoSurf), session load/save (SessionLoad),
// and an authentication gate for admin routes (Auth).
package main

import (
	"net/http"

	"github.com/bensabler/milos-residence/internal/helpers"
	"github.com/justinas/nosurf"
)

// NoSurf applies CSRF protection to the downstream handler chain using nosurf.
// It sets a secure, HttpOnly base cookie and enforces token validation on
// state-changing requests.
//
// Parameters:
//   - next: the next http.Handler in the chain.
//
// Returns:
//   - http.Handler: a handler that validates CSRF tokens on incoming requests.
//
// Notes:
//   - Cookie.Secure is bound to app.InProduction to avoid HTTPS-only cookies
//     in local development.
//   - SameSite Lax is a safe default that defends most CSRF vectors while
//     keeping top-level POST redirects functional.
func NoSurf(next http.Handler) http.Handler {
	// Wrap the next handler with nosurf’s token generation/verification.
	csrfHandler := nosurf.New(next)

	// Establish cookie policy for the CSRF base cookie.
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,                 // prevent JavaScript access
		Path:     "/",                  // send with all requests
		Secure:   app.InProduction,     // HTTPS-only in production
		SameSite: http.SameSiteLaxMode, // sensible CSRF default
	})

	return csrfHandler
}

// SessionLoad loads the session for the request and ensures it is saved
// after the downstream handler completes. This enables handlers to read
// and write session data without manual lifecycle management.
//
// Parameters:
//   - next: the next http.Handler in the chain.
//
// Returns:
//   - http.Handler: a handler that manages session load/save for each request.
func SessionLoad(next http.Handler) http.Handler {
	// scs wraps the handler to hydrate and persist session state.
	return session.LoadAndSave(next)
}

// Auth enforces that the caller is authenticated (has "user_id" in session)
// before allowing access to protected routes. Unauthenticated users are
// redirected to the login page with a one-time error message.
//
// Parameters:
//   - next: the protected handler to run after authentication succeeds.
//
// Returns:
//   - http.Handler: a handler that redirects unauthenticated users to /user/login.
//
// Side effects:
//   - Sets a session error flash message: "Log in first!"
//   - Issues an HTTP 303 See Other redirect on failure.
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Gate access based on session authentication marker.
		if !helpers.IsAuthenticated(r) {
			// Let the UI show a concise reason for the redirect.
			session.Put(r.Context(), "error", "Log in first!")

			// Redirect to the login page using a safe 303 response.
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}

		// Forward to the protected handler when authenticated.
		next.ServeHTTP(w, r)
	})
}
