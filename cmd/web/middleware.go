// Package main (middleware) provides HTTP middleware for CSRF protection and session loading.
package main

import (
	"net/http"

	"github.com/justinas/nosurf"
)

// NoSurf wraps a handler with CSRF protection for all POST requests using nosurf.
// It sets a base cookie configured for HttpOnly and SameSite=Lax. In production,
// ensure app.InProduction is true so the cookie is marked Secure.
func NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   app.InProduction,
		SameSite: http.SameSiteLaxMode,
	})

	return csrfHandler
}

// SessionLoad loads and saves the session on every request.
// Place this middleware high in the chain so downstream handlers can access session data.
func SessionLoad(next http.Handler) http.Handler { return session.LoadAndSave(next) }
