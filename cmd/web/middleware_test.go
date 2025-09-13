// Command web middleware tests ensure wrappers return the correct handler types
// so they can be composed reliably in the router.
package main

import (
	"net/http"
	"testing"
)

// TestNoSurf asserts that NoSurf returns an http.Handler wrapper compatible
// with chi middleware chains.
func TestNoSurf(t *testing.T) {
	var myH myHandler

	// Wrap a no-op handler and verify the resulting type.
	h := NoSurf(&myH)
	switch v := h.(type) {
	case http.Handler:
		// ok
	default:
		t.Errorf("type is not http.Handler, but is %T", v)
	}
}

// TestSessionLoad asserts that SessionLoad returns an http.Handler wrapper
// that can be composed in the middleware pipeline.
func TestSessionLoad(t *testing.T) {
	var myH myHandler

	// Wrap a no-op handler and verify the resulting type.
	h := SessionLoad(&myH)
	switch v := h.(type) {
	case http.Handler:
		// ok
	default:
		t.Errorf("type is not http.Handler, but is %T", v)
	}
}
