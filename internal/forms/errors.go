package forms

// errors stores validation messages keyed by form field name.
type errors map[string][]string

// Add appends a human-readable message for the specified field.
func (e errors) Add(field, message string) {
	// get the current list of messages for this field (or nil if none yet)
	// then append the new message to the slice
	e[field] = append(e[field], message)
}

// Get returns the first message for a field, or "" if none exist.
func (e errors) Get(field string) string {
	// look up all messages recorded for the field
	es := e[field]

	// if no messages were recorded, return an empty string for easy checks
	if len(es) == 0 {
		return ""
	}

	// otherwise, return the first message for concise inline display
	return es[0]
}
