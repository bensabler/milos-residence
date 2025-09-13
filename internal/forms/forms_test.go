package forms

import (
	"net/http/httptest"
	"net/url"
	"testing"
)

// TestForm_Valid verifies that a new, untouched form is considered valid.
func TestForm_Valid(t *testing.T) {
	form := New(url.Values{})
	if !form.Valid() {
		t.Error("got invalid when form should be valid (no errors)")
	}
}

// TestForm_Required verifies Required() records errors for missing fields
// and clears when all required fields are present.
func TestForm_Required(t *testing.T) {
	form := New(url.Values{})
	form.Required("a", "b", "c")
	if form.Valid() {
		t.Error("form shows valid when required fields are missing")
	}

	posted := url.Values{}
	posted.Add("a", "A")
	posted.Add("b", "B")
	posted.Add("c", "C")
	form = New(posted)
	form.Required("a", "b", "c")
	if !form.Valid() {
		t.Error("form reports missing required fields when all are present")
	}
}

// TestForm_Has verifies Has() returns false and records an error for a blank
// field, and true when the field has a value.
func TestForm_Has(t *testing.T) {
	r := httptest.NewRequest("POST", "/whatever", nil)
	form := New(r.PostForm)

	has := form.Has("whatever")
	if has {
		t.Error("form shows has field when it does not")
	}

	postedData := url.Values{}
	postedData.Add("a", "a")
	form = New(postedData)

	has = form.Has("a")
	if !has {
		t.Error("shows form does not have field when it should")
	}
}

// TestForm_MinLength ensures MinLength() flags too-short values and passes when
// the minimum length requirement is satisfied.
func TestForm_MinLength(t *testing.T) {
	r := httptest.NewRequest("POST", "/whatever", nil)
	form := New(r.PostForm)
	form.MinLength("x", 10)
	if form.Valid() {
		t.Error("form shows min length for non-existent field")
	}
	isError := form.Errors.Get("x")
	if isError == "" {
		t.Error("should have an error, but did not get one")
	}

	postedValues := url.Values{}
	postedValues.Add("some field", "some value")
	form = New(postedValues)
	form.MinLength("some_field", 100) // note: different field name than posted

	if form.Valid() {
		t.Error("shows minlength of 100 when data is shorter")
	}

	postedValues = url.Values{}
	postedValues.Add("another_field", "abc123")
	form = New(postedValues)
	form.MinLength("another_field", 1)

	if !form.Valid() {
		t.Error("shows minlength of 1 is not met when it is")
	}
	isError = form.Errors.Get("another_field")
	if isError != "" {
		t.Error("should not have an error, but got one")
	}
}

// TestForm_IsEmail validates that IsEmail() fails for empty/invalid values
// and passes for syntactically valid email addresses.
func TestForm_IsEmail(t *testing.T) {
	postedValues := url.Values{}
	form := New(postedValues)
	form.IsEmail("x")
	if form.Valid() {
		t.Error("form shows valid email for non-existent field")
	}

	postedValues = url.Values{}
	postedValues.Add("email", "me@here.com")
	form = New(postedValues)
	form.IsEmail("email")
	if !form.Valid() {
		t.Error("got an invalid email when we should not have")
	}

	postedValues = url.Values{}
	postedValues.Add("email", "x")
	form = New(postedValues)
	form.IsEmail("email")
	if form.Valid() {
		t.Error("got a valid email for an invalid email")
	}
}
