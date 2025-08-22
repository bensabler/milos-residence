package main

import (
	"net/http"
	"testing"
)

// TestNoSurf confirms NoSurf(...) returns a value that implements http.Handler.
func TestNoSurf(t *testing.T) {
	// create a dummy downstream handler to wrap
	var myH myHandler

	// apply the CSRF middleware to the dummy handler
	h := NoSurf(&myH)

	// verify the wrapped handler satisfies http.Handler
	switch v := h.(type) {
	case http.Handler:
		// success: the middleware returns an http.Handler
		// do nothing
	default:
		// mismatch: report the unexpected type
		t.Errorf("type is not http.Handler, but is %t", v)
	}
}

// TestSessionLoad confirms SessionLoad(...) returns an http.Handler wrapper.
func TestSessionLoad(t *testing.T) {
	// create a dummy downstream handler to wrap
	var myH myHandler

	// wrap the handler with session load/save middleware
	h := SessionLoad(&myH)

	// verify the result implements http.Handler for the middleware chain
	switch v := h.(type) {
	case http.Handler:
		// success: the middleware returns an http.Handler
		// do nothing
	default:
		// failure: the result did not implement http.Handler
		t.Errorf("type is not http.Handler, but is %t", v)
	}
}
