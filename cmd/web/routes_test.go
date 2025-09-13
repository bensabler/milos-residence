// Command web routing tests validate router construction and ensure the routes()
// function returns a chi.Mux suitable for http.Server.
package main

import (
	"testing"

	"github.com/bensabler/milos-residence/internal/config"
	"github.com/go-chi/chi/v5"
)

// TestRoutes confirms that routes() returns a *chi.Mux, the expected router type.
// This guards against accidental return type changes that would break server wiring.
func TestRoutes(t *testing.T) {
	var app config.AppConfig

	// Build the router from the provided (test) AppConfig.
	mux := routes(&app)

	// Type assertion: expect a chi mux router.
	switch v := mux.(type) {
	case *chi.Mux:
		// ok
	default:
		t.Errorf("type is not *chi.Mux, but is %T", v)
	}
}
