package main

import (
	"net/http"

	"github.com/justinas/nosurf"
)

// NoSurf wraps the next handler with CSRF protection for unsafe methods.
func NoSurf(next http.Handler) http.Handler {
	// create a CSRF-aware handler around the downstream handler
	csrfHandler := nosurf.New(next)

	// configure the cookie that carries the CSRF token
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,                 // prevent JavaScript from reading the token
		Path:     "/",                  // send the cookie for all paths
		Secure:   app.InProduction,     // HTTPS-only in production
		SameSite: http.SameSiteLaxMode, // safe default while still allowing normal nav
	})

	// return the wrapped handler so requests pass through CSRF checks
	return csrfHandler
}

// SessionLoad ensures the session is loaded before the handler runs and saved after.
func SessionLoad(next http.Handler) http.Handler {
	// wrap the downstream handler with session load/save middleware
	return session.LoadAndSave(next)
}
