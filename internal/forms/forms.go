package forms

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/asaskevich/govalidator"
)

// Form wraps url.Values with an errors accumulator used during validation.
// Call validation helpers (Required, MinLength, etc.) and then check Valid.
type Form struct {
	url.Values
	Errors errors
}

// Valid reports whether any validation errors were recorded.
// A Form is valid when no errors have been added.
func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}

// New constructs a Form from provided url.Values.
// The returned Form has an initialized errors map.
func New(data url.Values) *Form {
	return &Form{
		data,
		errors(map[string][]string{}),
	}
}

// Required asserts that each named field is present and non-blank.
// On failure, a "cannot be blank" error is recorded for the field.
func (f *Form) Required(fields ...string) {
	for _, field := range fields {
		value := f.Get(field)
		if strings.TrimSpace(value) == "" {
			f.Errors.Add(field, "This field cannot be blank")
		}
	}
}

// Has reports whether field exists with a non-empty value.
// Side effect: if blank, an error is recorded for the field and the result is false.
func (f *Form) Has(field string) bool {
	x := f.Get(field)
	if x == "" {
		f.Errors.Add(field, "This field cannot be blank")
		return false
	}
	return true
}

// MinLength asserts that field's value length is at least length characters.
// On failure, an error is recorded and the result is false.
func (f *Form) MinLength(field string, length int) bool {
	x := f.Get(field)
	if len(x) < length {
		f.Errors.Add(field, fmt.Sprintf("This field must be at least %d characters long", length))
		return false
	}
	return true
}

// IsEmail asserts that field contains a syntactically valid email address.
// Uses govalidator.IsEmail for format validation; records an error on failure.
func (f *Form) IsEmail(field string) {
	if !govalidator.IsEmail(f.Get(field)) {
		f.Errors.Add(field, "Invalid email address")
	}
}
