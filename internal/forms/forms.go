// Package forms centralizes request form validation with a small API designed
// for handlers and templates. It wraps url.Values, accumulates errors, and
// exposes helpers like Required, MinLength, and IsEmail.
package forms

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/asaskevich/govalidator"
)

// Form wraps url.Values with an errors accumulator used during validation.
// Typical flow: f := New(r.PostForm); f.Required("name"); if !f.Valid() { ... }.
type Form struct {
	url.Values
	Errors errors
}

// Valid reports whether any validation errors were recorded.
// Returns true when no errors have been added.
func (f *Form) Valid() bool {
	// A form is valid when its error map is empty.
	return len(f.Errors) == 0
}

// New constructs a Form from provided url.Values with an initialized error map.
// Usage: f := New(r.PostForm)
func New(data url.Values) *Form {
	// Initialize with the provided data and a fresh errors map.
	return &Form{
		Values: data,
		Errors: errors(map[string][]string{}),
	}
}

// Required asserts that each named field is present and non-blank.
// On failure, adds "This field cannot be blank" for each missing field.
// Usage: f.Required("first_name", "email")
func (f *Form) Required(fields ...string) {
	for _, field := range fields {
		// Trim whitespace to catch values like "  ".
		value := f.Get(field)
		if strings.TrimSpace(value) == "" {
			f.Errors.Add(field, "This field cannot be blank")
		}
	}
}

// Has reports whether field exists with a non-empty value.
// Side effect: records an error and returns false when blank.
// Usage: if !f.Has("email") { ... }
func (f *Form) Has(field string) bool {
	// Read once; an empty string means missing or blank.
	x := f.Get(field)
	if x == "" {
		f.Errors.Add(field, "This field cannot be blank")
		return false
	}
	return true
}

// MinLength asserts that field's value length is at least length characters.
// Returns false and records an error when the requirement is not met.
// Usage: if !f.MinLength("password", 8) { ... }
func (f *Form) MinLength(field string, length int) bool {
	// Pull the raw value so templates/handlers can reflect back the user's input.
	x := f.Get(field)
	if len(x) < length {
		f.Errors.Add(field, fmt.Sprintf("This field must be at least %d characters long", length))
		return false
	}
	return true
}

// IsEmail asserts that field contains a syntactically valid email address.
// Uses govalidator.IsEmail for format validation; records an error on failure.
// Usage: f.IsEmail("email")
func (f *Form) IsEmail(field string) {
	// Delegate to a vetted library for basic syntax checks.
	if !govalidator.IsEmail(f.Get(field)) {
		f.Errors.Add(field, "Invalid email address")
	}
}
