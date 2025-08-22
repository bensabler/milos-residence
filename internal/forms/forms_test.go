package forms

import (
	"net/http/httptest"
	"net/url"
	"testing"
)

// TestForm_Valid verifies that a brand-new form with no recorded errors is valid.
func TestForm_Valid(t *testing.T) {
	// build an empty form (no fields, no errors)
	form := New(url.Values{})

	// expect Valid() to be true because no validators added errors
	if !form.Valid() {
		t.Error("got invalid when form should be valid (no errors)")
	}
}

// TestForm_Required checks that Required() flags missing fields and passes when present.
func TestForm_Required(t *testing.T) {
	// start with an empty form and require three fields
	form := New(url.Values{})
	form.Required("a", "b", "c")

	// since the fields are missing, the form should be invalid
	if form.Valid() {
		t.Error("form shows valid when required fields are missing")
	}

	// now provide values for all required fields
	posted := url.Values{}
	posted.Add("a", "A")
	posted.Add("b", "B")
	posted.Add("c", "C")

	// rebuild the form with posted data and require the same fields
	form = New(posted)
	form.Required("a", "b", "c")

	// with all fields present (non-blank), the form should be valid
	if !form.Valid() {
		t.Error("form reports missing required fields when all are present")
	}
}

// TestForm_Has ensures Has() correctly reports presence/absence of a single field.
func TestForm_Has(t *testing.T) {
	// create a request without any posted data
	r := httptest.NewRequest("POST", "/whatever", nil)
	form := New(r.PostForm)

	// ask for a field that doesn't exist; expect false
	has := form.Has("whatever")
	if has {
		t.Error("form shows has field when it does not")
	}

	// now create posted data with one key/value
	postedData := url.Values{}
	postedData.Add("a", "a")
	form = New(postedData)

	// ask for the existing key; expect true
	has = form.Has("a")
	if !has {
		t.Error("shows form does not have field when it should")
	}
}

// TestForm_MinLength validates length constraints for present and missing fields.
func TestForm_MinLength(t *testing.T) {
	// start with no posted data
	r := httptest.NewRequest("POST", "/whatever", nil)
	form := New(r.PostForm)

	// checking min length on a missing field should trigger an error and invalid form
	form.MinLength("x", 10)
	if form.Valid() {
		t.Error("form shows min length for non-existent field")
	}

	// confirm an error message exists for the missing field
	isError := form.Errors.Get("x")
	if isError == "" {
		t.Error("should have an error, but did not get one")
	}

	// provide one field/value pair
	postedValues := url.Values{}
	postedValues.Add("some field", "some value")
	form = New(postedValues)

	// check a different field name (underscore vs. space) so it should behave as missing
	form.MinLength("some_field", 100)
	if form.Valid() {
		t.Error("shows minlength of 100 when data is shorter")
	}

	// use a field that exists with a short min-length requirement that should pass
	postedValues = url.Values{}
	postedValues.Add("another_field", "abc123")
	form = New(postedValues)

	form.MinLength("another_field", 1)
	if !form.Valid() {
		t.Error("shows minlength of 1 is not met when it is")
	}

	// ensure no error is recorded for the passing case
	isError = form.Errors.Get("another_field")
	if isError != "" {
		t.Error("should have an error, but did not get one")
	}
}

// TestForm_IsEmail verifies email format validation for missing, valid, and invalid cases.
func TestForm_IsEmail(t *testing.T) {
	// no email field present: should be treated as invalid format
	postedValues := url.Values{}
	form := New(postedValues)

	form.IsEmail("x")
	if form.Valid() {
		t.Error("form shows valid email for non-existent field")
	}

	// a syntactically valid email should pass
	postedValues = url.Values{}
	postedValues.Add("email", "me@here.com")
	form = New(postedValues)

	form.IsEmail("email")
	if !form.Valid() {
		t.Error("got an invalid email when we should not have")
	}

	// an invalid email string should fail
	postedValues = url.Values{}
	postedValues.Add("email", "x")
	form = New(postedValues)

	form.IsEmail("email")
	if form.Valid() {
		t.Error("got a valid email for an invalid email")
	}
}
