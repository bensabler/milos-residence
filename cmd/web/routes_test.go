package main

import (
	"testing"

	"github.com/bensabler/milos-residence/internal/config"
	"github.com/go-chi/chi/v5"
)

// TestRoutes ensures routes(...) builds and returns a chi router instance.
func TestRoutes(t *testing.T) {
	// create a minimal app config for the router to attach to
	var app config.AppConfig

	// build the application's top-level router
	mux := routes(&app)

	// check the concrete type to confirm we're returning a chi mux/router
	switch v := mux.(type) {
	case *chi.Mux:
		// success: the router is the expected type
		// do nothing
	default:
		// not the expected router type; report the mismatch
		t.Errorf("type is not *chi.Mux, but is %t", v)
	}
}
