package main

import (
	"testing"

	"github.com/bensabler/milos-residence/internal/config"
	"github.com/go-chi/chi/v5"
)

func TestRoutes(t *testing.T) {
	var app config.AppConfig

	mux := routes(&app)

	switch v := mux.(type) {
	case *chi.Mux:
		// do notihing; test passed
	default:
		t.Errorf("type is not *chi.Mux, but is %t", v)
	}
}