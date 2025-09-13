// Package forms provides lightweight form validation helpers for HTTP handlers.
// It wraps url.Values with a Form type and an errors collector to record and
// report validation failures in a way that slots cleanly into templates.
package forms

// errors stores validation messages keyed by field name.
// A field may accumulate multiple messages (e.g., required + min length).
type errors map[string][]string

// Add appends a validation message for the given field.
// Use for recording one or more errors during validation.
func (e errors) Add(field, message string) {
	e[field] = append(e[field], message)
}

// Get returns the first error message for field, or the empty string if none.
// This is convenient for templates that display a single message per field.
func (e errors) Get(field string) string {
	es := e[field]
	if len(es) == 0 {
		return ""
	}
	return es[0]
}
