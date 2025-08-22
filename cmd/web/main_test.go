package main

import "testing"

// TestRun verifies that run() completes initialization without returning an error.
func TestRun(t *testing.T) {
	// call the app initialization function
	err := run()

	// if run() returns a non-nil error, the bootstrapping failed
	if err != nil {
		t.Error("Failed run()")
	}
}
