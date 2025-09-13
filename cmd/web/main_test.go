// Command web tests cover startup/bootstrap routines for the web binary.
// This file verifies that run() completes without returning an error.
package main

import "testing"

// TestRun validates that run() performs application bootstrap successfully.
// It expects no error on normal test initialization.
func TestRun(t *testing.T) {
	// Execute the bootstrap routine and assert success.
	_, err := run()
	if err != nil {
		t.Error("Failed run()")
	}
}
