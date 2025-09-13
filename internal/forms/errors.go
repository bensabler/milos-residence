// Package forms provides lightweight form validation helpers for HTTP handlers.
// It wraps url.Values with a Form type and an errors collector to record and
// report validation failures in a way that slots cleanly into templates.
package forms

// errors stores validation messages keyed by field name.
// A field may accumulate multiple messages (e.g., required + min length).
type errors map[string][]string

// Add appends a validation message for the given field.
// Usage: e.Add("email", "Invalid email address")
func (e errors) Add(field, message string) {
	// Append message to the slice for this field; initialize-on-write semantics.
	e[field] = append(e[field], message)
}

// Get returns the first error message for field, or the empty string if none.
// Templates commonly show only the first message per field.
func (e errors) Get(field string) string {
	// Read the slice and return the first entry if present.
	es := e[field]
	if len(es) == 0 {
		return ""
	}
	return es[0]
}
