package main

import (
	"net/http"
	"os"
	"testing"
)

// TestMain is the entry point for this package's tests.
func TestMain(m *testing.M) {
	// run the entire test suite for this package and exit with its status code
	os.Exit(m.Run())
}

// myHandler is a minimal stub that satisfies http.Handler for middleware tests.
type myHandler struct{}

// ServeHTTP implements http.Handler; it's intentionally a no-op for testing.
func (mh *myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// no-op: used as a stand-in handler so tests can wrap it with middleware
}
