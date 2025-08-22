package forms

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/asaskevich/govalidator"
)

// Form wraps url.Values and a per-field error list used during validation.
type Form struct {
	url.Values
	Errors errors
}

// Valid reports true when no validation errors have been recorded.
func (f *Form) Valid() bool {
	// if the error map is empty, the form has passed all current checks
	return len(f.Errors) == 0
}

// New constructs a Form backed by the provided url.Values.
func New(data url.Values) *Form {
	// create a new form to carry inputs and any validation messages
	return &Form{
		// keep the original values so validators and templates can read inputs
		data,
		// start with an empty error list per field
		errors(map[string][]string{}),
	}
	// TODO: if callers might mutate 'data' after New, consider copying first.
}

// Required ensures each named field is present and not just whitespace.
func (f *Form) Required(fields ...string) {
	// iterate over every field the caller marked as required
	for _, field := range fields {
		// read what the user provided for this field ("" if missing)
		value := f.Get(field)

		// trim spaces so strings like "   " don’t pass as “filled in”
		if strings.TrimSpace(value) == "" {
			// record a helpful message that templates can show next to the field
			f.Errors.Add(field, "This field cannot be blank")
			// keep looping so we can report all missing fields at once
		}
	}
}

// Has reports whether a specific field exists and isn't empty.
func (f *Form) Has(field string) bool {
	// store the field value in x
	x := f.Get(field)

	// if x is an empty string
	if x == "" {
		// add the field and an error message, then return
		f.Errors.Add(field, "This field cannot be blank")
		return false
	}

	// otherwise return
	return true
}

// MinLength checks that a field's value has at least length characters.
func (f *Form) MinLength(field string, length int) bool {
	// read the value provided for this field
	x := f.Get(field)

	// if the value is shorter than the required length
	if len(x) < length {
		// add an actionable message for the user and fail the check
		f.Errors.Add(field, fmt.Sprintf("This field must be at least %d characters long", length))
		return false
	}

	// the value meets the length requirement
	return true
	// TODO: consider trimming before measuring if leading/trailing spaces shouldn't count.
}

// IsEmail adds an error if the field does not contain a syntactically valid email.
func (f *Form) IsEmail(field string) {
	// let the validator library check email syntax for us
	if !govalidator.IsEmail(f.Get(field)) {
		// record a clear message that templates can show inline
		f.Errors.Add(field, "Invalid email address")
	}
	// note: this checks format only; it does not verify deliverability.
}
