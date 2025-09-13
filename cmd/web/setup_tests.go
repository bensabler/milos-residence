// Command web test scaffolding provides shared test setup and small stubs used
// by multiple test files in this package.
package main

import (
	"net/http"
	"os"
	"testing"
)

// TestMain provides a central entry point for tests in this package.
// It currently delegates directly to m.Run() without additional setup.
func TestMain(m *testing.M) {
	// Execute the test suite and exit with its status code.
	os.Exit(m.Run())
}

// myHandler is a minimal http.Handler stub for middleware tests.
// It records no state and writes no response by design.
type myHandler struct{}

// ServeHTTP satisfies http.Handler for myHandler and performs no operation.
// It allows middleware wrappers to be tested in isolation.
func (mh *myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// no-op
}
