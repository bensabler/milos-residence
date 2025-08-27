package forms

// errors implements the Error Collector pattern for validation message management.
// This type provides a structured way to collect, organize, and retrieve validation
// error messages associated with specific form fields. It demonstrates how Go's type
// system can be used to create domain-specific collections that provide both type
// safety and convenient methods for common operations.
//
// The Error Collector pattern is essential for user-friendly web applications because:
// 1. Users need to see which specific fields have problems, not just "validation failed"
// 2. Multiple validation errors per field should be collected and presented together
// 3. Templates need easy access to field-specific errors for inline display
// 4. Error messages should be localized and user-friendly, not technical
// 5. The same validation errors may need to be displayed in different formats
//
// By using a map structure keyed by field name, this implementation provides O(1)
// lookup performance for error retrieval while supporting multiple error messages
// per field through slice values. This design scales efficiently from simple forms
// with a few fields to complex forms with dozens of validation rules.
//
// Design Pattern: Error Collector - structured collection of validation failures
// Design Pattern: Domain-Specific Collection - specialized map type with convenience methods
// Design Pattern: Multi-Value Map - supports multiple error messages per field
type errors map[string][]string

// Add implements the Error Accumulator pattern for collecting validation failures.
// This method appends new error messages to the collection associated with a specific
// form field, enabling multiple validation rules to report failures for the same field
// without overwriting previous error messages. It demonstrates how validation frameworks
// can provide comprehensive feedback to users about all problems that need fixing.
//
// The accumulator approach is crucial for good user experience because it prevents
// the frustrating cycle where users fix one validation error only to discover another
// error for the same field. By collecting all validation failures at once, users
// can see the complete picture of what needs to be corrected.
//
// Design Pattern: Error Accumulator - collects multiple errors without overwriting
// Design Pattern: Append-Only Collection - preserves all error messages for comprehensive feedback
// Parameters:
//
//	field: The form field name that failed validation (e.g., "email", "first_name")
//	message: User-friendly error message describing what needs to be fixed
func (e errors) Add(field, message string) {
	// Retrieve the current list of error messages for this field
	// If this is the first error for this field, the map lookup returns nil
	// which is the zero value for a slice and is safe to append to
	// This demonstrates Go's zero-value-is-useful design principle

	// Append the new error message to the existing error list for this field
	// The append function handles both cases automatically:
	// 1. If e[field] is nil (first error), it creates a new slice with the message
	// 2. If e[field] contains existing errors, it appends to the existing slice
	// This dual behavior eliminates the need for explicit initialization checking
	e[field] = append(e[field], message)

	// The result is that each field can accumulate multiple error messages
	// which can then be displayed together in templates or processed by handlers
	// This enables comprehensive validation feedback without losing information
}

// Get implements the Error Retrieval pattern for accessing field-specific error messages.
// This method provides convenient access to the first (typically most important)
// error message for a specific form field. It handles the common case where templates
// or handlers need to display a single error message inline with form fields,
// while gracefully handling cases where no errors exist for the requested field.
//
// The "first error only" approach balances comprehensive error reporting with
// clean user interface design. While all errors are collected and available,
// displaying only the first error per field keeps form layouts manageable and
// prevents overwhelming users with too much information at once.
//
// Design Pattern: Error Retrieval - convenient access to collected error messages
// Design Pattern: Safe Navigation - graceful handling of missing data without panics
// Design Pattern: First Match - returns most relevant error when multiple exist
// Parameters:
//
//	field: The form field name to retrieve error messages for
//
// Returns: The first error message for the field, or empty string if no errors exist
func (e errors) Get(field string) string {
	// Look up all error messages recorded for the specified field
	// This demonstrates safe map access patterns in Go - missing keys return
	// the zero value (nil for slices) rather than panicking or requiring
	// explicit existence checking before access
	es := e[field]

	// Check if any error messages were recorded for this field
	// A nil slice or empty slice both have length 0, so this check handles
	// both the "field never validated" and "field validated successfully" cases
	if len(es) == 0 {
		// No errors recorded for this field - return empty string for safe template usage
		// Empty string is the natural zero value for error display and can be safely
		// used in templates without additional null checking or conditional logic
		return ""
	}

	// Return the first error message for concise inline display
	// Index 0 is safe here because we've already confirmed len(es) > 0
	// The first error is typically the most fundamental validation failure
	// (e.g., "required" errors usually come before format validation errors)
	return es[0]

	// Note: While only the first error is returned by Get(), all error messages
	// remain available in the errors collection. Advanced templates or handlers
	// can access e[field] directly to retrieve all error messages for a field
	// if more comprehensive error display is needed for specific use cases
}
