package forms

import (
	"net/http/httptest"
	"net/url"
	"testing"
)

// TestForm_Valid implements the State Testing pattern for form validation status verification.
// This test validates that the form validation system correctly identifies forms with no
// validation errors as valid, establishing the baseline behavior that enables form processing
// workflows to make correct decisions about whether to proceed with business operations or
// re-display forms with error messages for user correction.
//
// Testing the "valid" state is foundational because:
// 1. **Positive Path Verification**: Ensures successful form processing works correctly
// 2. **Baseline Establishment**: Confirms that empty error state correctly indicates validity
// 3. **Business Logic Foundation**: Form processing decisions depend on accurate validity checking
// 4. **User Experience**: Valid forms should proceed smoothly without unnecessary roadblocks
// 5. **Integration Testing**: Other components rely on accurate validity reporting for proper operation
//
// This test establishes the zero-error state as the definition of form validity, which is
// a fundamental assumption used throughout the form processing system for decision making.
//
// Design Pattern: State Testing - verifies correct state identification and reporting
// Design Pattern: Baseline Testing - establishes expected behavior for successful scenarios
// Design Pattern: Zero-Value Testing - tests behavior with empty/initial state
func TestForm_Valid(t *testing.T) {
	// Create empty form with no validation errors
	// New() with empty url.Values creates a form in its initial state,
	// representing a form that hasn't been validated yet or has passed all validations
	form := New(url.Values{})

	// Verify that form without validation errors reports as valid
	// This tests the fundamental assumption that absence of errors indicates validity,
	// which is the basis for all form processing decision logic throughout the application
	if !form.Valid() {
		// Form incorrectly reports as invalid when no errors exist
		// This would indicate a fundamental problem with the validation state logic
		// that could cause valid forms to be rejected or re-displayed unnecessarily
		t.Error("got invalid when form should be valid (no errors)")
	}
}

// TestForm_Required implements the Validation Testing pattern for presence validation verification.
// This test provides comprehensive coverage of the Required validation method, testing both
// failure scenarios (missing required fields) and success scenarios (all required fields present).
// It demonstrates how validation testing should cover both positive and negative cases to ensure
// robust form processing behavior under all user input conditions.
//
// Required field validation is critical for web applications because:
// 1. **Data Integrity**: Ensures essential information is collected before processing
// 2. **Business Rule Enforcement**: Implements business requirements about mandatory data
// 3. **User Experience**: Provides clear feedback about what information is needed
// 4. **Security**: Prevents processing of incomplete data that could cause system errors
// 5. **Compliance**: Supports regulatory requirements for information collection
//
// The test validates both the error detection (when fields are missing) and the success
// path (when all required fields are provided), ensuring complete validation behavior coverage.
//
// Design Pattern: Validation Testing - comprehensive testing of validation rule implementation
// Design Pattern: Positive/Negative Testing - tests both success and failure scenarios
// Design Pattern: State Verification - validates form state changes after validation operations
func TestForm_Required(t *testing.T) {
	// Test negative case: form with missing required fields should be invalid

	// Create empty form representing user submission with no field values
	// This simulates a user submitting a form without filling in any required fields,
	// which should trigger validation errors for all required fields
	form := New(url.Values{})

	// Apply required validation to multiple fields that are missing from the form
	// This tests the validation system's ability to check multiple fields simultaneously
	// and accumulate validation errors for comprehensive user feedback
	form.Required("a", "b", "c")

	// Verify that form with missing required fields correctly reports as invalid
	// This validates that Required validation properly detects missing fields
	// and updates the form's validation state to prevent processing incomplete data
	if form.Valid() {
		// Form incorrectly reports as valid when required fields are missing
		// This would allow incomplete forms to be processed, potentially causing
		// business logic errors or data integrity problems
		t.Error("form shows valid when required fields are missing")
	}

	// Test positive case: form with all required fields present should be valid

	// Create form data containing values for all required fields
	// This simulates a user properly filling in all required information
	// before submitting the form for processing
	posted := url.Values{}
	posted.Add("a", "A") // Provide value for first required field
	posted.Add("b", "B") // Provide value for second required field
	posted.Add("c", "C") // Provide value for third required field

	// Create new form instance with complete field data
	// This represents a properly filled form ready for validation
	form = New(posted)

	// Apply same required validation to fields that now have values
	// This tests that Required validation correctly recognizes present fields
	// and allows forms with complete data to pass validation
	form.Required("a", "b", "c")

	// Verify that form with all required fields correctly reports as valid
	// This validates the positive path where validation should succeed
	// and allow form processing to continue with business operations
	if !form.Valid() {
		// Form incorrectly reports as invalid when all required fields are present
		// This would prevent valid forms from being processed, creating poor user experience
		// and potentially blocking legitimate business operations
		t.Error("form reports missing required fields when all are present")
	}
}

// TestForm_Has implements the Individual Field Testing pattern for single field presence verification.
// This test focuses on the Has method which checks individual field presence and provides both
// boolean return values for programmatic use and error collection for user feedback. It demonstrates
// testing of methods that provide dual interfaces for different consumption patterns within the
// same application.
//
// Individual field testing complements bulk validation testing by:
// 1. **Granular Control**: Enables specific field checking in complex validation scenarios
// 2. **Conditional Logic**: Supports validation rules that depend on other field values
// 3. **Progressive Enhancement**: Allows incremental validation during user input
// 4. **Custom Workflows**: Enables specialized validation patterns for specific business needs
// 5. **Debugging Support**: Provides focused testing for specific validation problems
//
// The dual return pattern (boolean + error collection) enables the same method to support
// both programmatic decision making and user interface error display requirements.
//
// Design Pattern: Individual Field Testing - focused testing of single field validation
// Design Pattern: Dual Interface Testing - validates both return value and side effect behavior
// Design Pattern: Boolean Logic Testing - verifies correct true/false return values
func TestForm_Has(t *testing.T) {
	// Test negative case: missing field should return false and record error

	// Create HTTP request with no form data to simulate empty form submission
	// httptest.NewRequest provides controlled test environment for HTTP operations
	// without requiring actual HTTP server or network operations
	r := httptest.NewRequest("POST", "/whatever", nil)

	// Create form from empty request data
	// r.PostForm is empty when no form data is submitted, simulating user
	// submitting empty form or accessing form without providing required information
	form := New(r.PostForm)

	// Test Has method with field that doesn't exist in form submission
	// This simulates checking for required information that user didn't provide
	has := form.Has("whatever")

	// Verify that Has correctly returns false for missing field
	// Boolean return value enables programmatic decision making in handlers
	// while error collection provides user feedback for form re-display
	if has {
		// Method incorrectly returns true for missing field
		// This would allow incomplete forms to proceed when they should be rejected
		t.Error("form shows has field when it does not")
	}

	// Test positive case: present field should return true without errors

	// Create form data with specific field value present
	// This simulates user providing the requested information in form submission
	postedData := url.Values{}
	postedData.Add("a", "a") // Add field with value to test presence detection

	// Create form instance with field data present
	form = New(postedData)

	// Test Has method with field that exists in form data
	// This verifies that Has correctly identifies present fields and returns true
	has = form.Has("a")

	// Verify that Has correctly returns true for existing field
	// This validates the positive path where field presence should be confirmed
	if !has {
		// Method incorrectly returns false for existing field
		// This would prevent valid forms from being processed correctly
		t.Error("shows form does not have field when it should")
	}
}

// TestForm_MinLength implements the Length Validation Testing pattern for field content verification.
// This test provides comprehensive coverage of minimum length validation, including edge cases
// like missing fields, fields that are too short, and fields that meet length requirements.
// It demonstrates how validation testing should cover boundary conditions and error scenarios
// that commonly occur in production web applications with user-generated content.
//
// Length validation testing is essential because:
// 1. **Data Quality**: Ensures collected information meets business requirements for completeness
// 2. **Security**: Prevents submission of empty or meaningless data that could cause processing errors
// 3. **User Experience**: Provides clear feedback about content requirements and expectations
// 4. **Business Rules**: Enforces organizational policies about data collection standards
// 5. **Integration**: Supports downstream systems that expect data to meet length requirements
//
// The test covers multiple scenarios including missing fields, insufficient content, and valid content
// to ensure robust validation behavior under all realistic user input conditions.
//
// Design Pattern: Length Validation Testing - comprehensive testing of content length requirements
// Design Pattern: Boundary Testing - validates behavior at length requirement boundaries
// Design Pattern: Error Message Testing - verifies appropriate error feedback for user guidance
func TestForm_MinLength(t *testing.T) {
	// Test negative case: missing field should fail length validation

	// Create HTTP request with no form data
	r := httptest.NewRequest("POST", "/whatever", nil)
	form := New(r.PostForm)

	// Apply minimum length validation to field that doesn't exist
	// This tests validation behavior when required field is completely missing,
	// which should be handled gracefully with appropriate error messages
	form.MinLength("x", 10)

	// Verify that form with missing field correctly reports as invalid
	if form.Valid() {
		// Form incorrectly passes validation when field is missing
		// This would allow incomplete forms to be processed inappropriately
		t.Error("form shows min length for non-existent field")
	}

	// Verify that appropriate error message was recorded for missing field
	// Error messages guide users toward providing the required information
	isError := form.Errors.Get("x")
	if isError == "" {
		// No error message recorded for validation failure
		// Users wouldn't receive guidance about what needs to be corrected
		t.Error("should have an error, but did not get one")
	}

	// Test negative case: field too short should fail length validation

	// Create form data with field that doesn't meet length requirements
	postedValues := url.Values{}
	postedValues.Add("some field", "some value") // Add field with specific name
	form = New(postedValues)

	// Apply length validation to different field name (testing field name sensitivity)
	// This verifies that validation correctly identifies missing fields even when
	// other fields are present, demonstrating precise field-specific validation
	form.MinLength("some_field", 100) // Note: underscore vs space in field name

	// Verify that form with short field content correctly reports as invalid
	if form.Valid() {
		// Form incorrectly passes validation when field content is too short
		t.Error("shows minlength of 100 when data is shorter")
	}

	// Test positive case: field meeting length requirement should pass validation

	// Create form data with field that meets length requirements
	postedValues = url.Values{}
	postedValues.Add("another_field", "abc123") // 6 characters, should exceed minimum of 1
	form = New(postedValues)

	// Apply reasonable minimum length requirement that content should satisfy
	form.MinLength("another_field", 1)

	// Verify that form with adequate field length correctly reports as valid
	if !form.Valid() {
		// Form incorrectly fails validation when field meets length requirement
		t.Error("shows minlength of 1 is not met when it is")
	}

	// Verify that no error message was recorded for valid field
	// Successful validation should not produce error messages for display
	isError = form.Errors.Get("another_field")
	if isError != "" {
		// Error message recorded for field that passed validation
		// This could confuse users by showing errors for valid input
		t.Error("should not have an error, but got one")
	}
}

// TestForm_IsEmail implements the Format Validation Testing pattern for email address verification.
// This test provides comprehensive coverage of email format validation including missing fields,
// valid email addresses, and invalid email formats. It demonstrates how validation testing should
// verify integration with third-party validation libraries while ensuring consistent error handling
// and user feedback across different validation scenarios.
//
// Email validation testing is crucial because:
// 1. **Communication Requirements**: Invalid emails break notification and confirmation workflows
// 2. **Data Quality**: Email format errors are common user mistakes requiring clear feedback
// 3. **Security**: Malformed emails could potentially cause processing errors or security issues
// 4. **User Experience**: Clear validation feedback helps users correct common email format mistakes
// 5. **Business Operations**: Valid emails are essential for customer communication and support
//
// The test validates the integration with external validation libraries while ensuring that
// error handling and user feedback patterns remain consistent with other validation methods.
//
// Design Pattern: Format Validation Testing - verifies correct format checking implementation
// Design Pattern: Third-Party Integration Testing - validates external library integration
// Design Pattern: Error Handling Testing - ensures consistent error feedback across validation types
func TestForm_IsEmail(t *testing.T) {
	// Test negative case: missing email field should fail validation

	// Create empty form representing submission without email field
	postedValues := url.Values{}
	form := New(postedValues)

	// Apply email validation to field that doesn't exist
	// This tests validation behavior when required email field is missing,
	// ensuring that email validation handles missing data appropriately
	form.IsEmail("x")

	// Verify that form without email field correctly reports as invalid
	if form.Valid() {
		// Form incorrectly passes validation when email field is missing
		// This would allow forms to be processed without required contact information
		t.Error("form shows valid email for non-existent field")
	}

	// Test positive case: valid email format should pass validation

	// Create form data with properly formatted email address
	postedValues = url.Values{}
	postedValues.Add("email", "me@here.com") // Standard email format that should validate
	form = New(postedValues)

	// Apply email validation to field with valid email format
	// This tests that email validation correctly recognizes standard email formats
	// and allows properly formatted addresses to pass validation
	form.IsEmail("email")

	// Verify that form with valid email correctly reports as valid
	if !form.Valid() {
		// Form incorrectly fails validation for properly formatted email
		// This would prevent users with valid emails from submitting forms successfully
		t.Error("got an invalid email when we should not have")
	}

	// Test negative case: invalid email format should fail validation

	// Create form data with improperly formatted email address
	postedValues = url.Values{}
	postedValues.Add("email", "x") // Invalid email format that should be rejected
	form = New(postedValues)

	// Apply email validation to field with invalid email format
	// This tests that email validation correctly identifies malformed email addresses
	// and prevents processing of forms with invalid contact information
	form.IsEmail("email")

	// Verify that form with invalid email correctly reports as invalid
	if form.Valid() {
		// Form incorrectly passes validation for malformed email address
		// This would allow invalid contact information to be stored in the system
		t.Error("got a valid email for an invalid email")
	}
}
