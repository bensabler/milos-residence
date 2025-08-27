package forms

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/asaskevich/govalidator"
)

// Form implements the Validation Framework pattern for HTTP form processing and validation.
// This struct wraps Go's standard url.Values (which represents HTML form data) with
// comprehensive validation capabilities, error collection, and user-friendly feedback
// mechanisms. It demonstrates how to add sophisticated behavior to existing types through
// composition rather than inheritance, following Go's design principles.
//
// The Validation Framework pattern provides several critical benefits for web applications:
// 1. Centralized validation logic - all validation rules in one place, not scattered across handlers
// 2. Reusable validation components - same validation can be applied to different forms
// 3. User-friendly error messaging - structured error collection with field-specific messages
// 4. Fluent interface design - method chaining enables readable validation rule definition
// 5. Separation of concerns - handlers focus on business logic, forms handle validation
//
// This approach scales well from simple contact forms to complex multi-step workflows
// because the validation logic remains consistent and maintainable regardless of form complexity.
//
// Design Pattern: Validation Framework - centralized form validation and error handling
// Design Pattern: Decorator - adds validation behavior to url.Values through composition
// Design Pattern: Fluent Interface - enables method chaining for readable validation rules
type Form struct {
	// url.Values is embedded to provide direct access to form data
	// This composition approach means Form inherits all url.Values methods
	// (Get, Set, Add, Del, etc.) while adding validation capabilities on top.
	// Embedding demonstrates Go's preference for composition over inheritance
	url.Values

	// Errors collects validation failures organized by field name
	// This enables templates to display field-specific error messages next to
	// the corresponding form inputs, providing clear user feedback about what
	// needs to be corrected. The errors type is defined in errors.go
	Errors errors
}

// Valid implements the Validation Result pattern for form validation status checking.
// This method provides a simple boolean interface for determining whether a form
// has passed all validation rules, enabling clean conditional logic in handlers
// without requiring detailed knowledge of the error collection implementation.
//
// The method demonstrates the Tell, Don't Ask principle - instead of exposing
// the internal error collection for handlers to examine, it provides a clear
// behavioral interface that encapsulates the validation state logic.
//
// Design Pattern: Validation Result - provides simple interface for validation status
// Design Pattern: Information Hiding - encapsulates error collection implementation details
// Returns: true if form has no validation errors, false if any validation rules failed
func (f *Form) Valid() bool {
	// Check if any validation errors have been recorded
	// An empty error collection (length 0) indicates all validation rules passed
	// This approach treats absence of errors as success rather than requiring
	// explicit success flags, following the zero-value-is-useful principle
	return len(f.Errors) == 0
}

// New implements the Factory pattern for Form instance creation with proper initialization.
// This factory function creates a Form with all necessary components initialized correctly,
// ensuring that the validation framework is ready for use without requiring complex
// setup code in handlers. It demonstrates how Go constructors should handle initialization
// of composite types with multiple components.
//
// The Factory pattern is particularly important here because Form contains multiple
// components (embedded url.Values and error collection) that must be initialized
// in a coordinated way to prevent nil pointer panics and ensure proper behavior.
//
// Design Pattern: Factory Method - creates properly initialized Form instances
// Design Pattern: Constructor Pattern - handles complex initialization logic
// Parameters:
//
//	data: Form data from HTTP request (typically r.PostForm or r.Form)
//
// Returns: Fully initialized Form ready for validation and error collection
func New(data url.Values) *Form {
	// Create Form instance with all components properly initialized
	// This constructor ensures both the embedded url.Values and the error
	// collection are ready for use without additional setup required by callers
	return &Form{
		// Store the form data for validation and template access
		// The form data contains all field values submitted by the user,
		// enabling validation rules to examine actual submitted values
		data,

		// Initialize empty error collection for validation failure messages
		// Starting with an empty map means Valid() will return true initially,
		// and validation methods can add errors as they discover problems
		errors(map[string][]string{}),
	}
}

// Required implements the Strategy pattern for presence validation of form fields.
// This validation strategy checks that specified form fields are present and contain
// non-empty values after whitespace trimming. It demonstrates how validation rules
// can be applied to multiple fields efficiently while collecting comprehensive
// error messages for user feedback.
//
// The Strategy pattern allows different validation algorithms (required, length,
// format, etc.) to be applied to the same form data structure, with each strategy
// focusing on a specific validation concern while maintaining consistent interfaces.
//
// Design Pattern: Strategy - implements one specific validation algorithm
// Design Pattern: Fluent Interface - enables chaining with other validation methods
// Design Pattern: Error Collector - accumulates validation failures for batch feedback
// Parameters:
//
//	fields: Variable number of field names that must be present and non-empty
func (f *Form) Required(fields ...string) {
	// Iterate through each field that should contain required data
	// Using variadic parameters allows this method to validate any number of
	// required fields in a single call, reducing code duplication in handlers
	for _, field := range fields {
		// Retrieve the submitted value for this field
		// Get() returns an empty string if the field was not submitted,
		// which is the standard behavior for HTML forms and url.Values
		value := f.Get(field)

		// Normalize the value by removing leading and trailing whitespace
		// TrimSpace handles all Unicode whitespace characters (spaces, tabs, newlines)
		// This prevents users from "cheating" required field validation by submitting
		// only whitespace characters, which would appear empty but technically have content
		if strings.TrimSpace(value) == "" {
			// Record a user-friendly error message for this field
			// The error message should be actionable and specific enough for users
			// to understand exactly what they need to do to fix the problem
			f.Errors.Add(field, "This field cannot be blank")

			// Continue checking other fields rather than stopping at first error
			// This provides better user experience by showing all validation problems
			// at once rather than forcing users to fix errors one at a time
		}
	}
}

// Has implements the Strategy pattern for field presence validation with error collection.
// This method checks whether a specific field exists in the form submission and contains
// non-empty content. Unlike Required (which validates multiple fields), Has focuses on
// a single field and returns a boolean result while also collecting error messages.
//
// This dual behavior (boolean return + error collection) enables both programmatic
// validation logic in handlers and user feedback through the error collection system.
//
// Design Pattern: Strategy - implements field presence validation algorithm
// Design Pattern: Dual Interface - provides both boolean result and error collection
// Parameters:
//
//	field: Name of the form field to check for presence and content
//
// Returns: true if field exists and has non-empty content, false otherwise
func (f *Form) Has(field string) bool {
	// Retrieve the submitted value for the specified field
	// This demonstrates consistent field access patterns across validation methods
	x := f.Get(field)

	// Check for empty field content and handle appropriately
	if x == "" {
		// Record error message for template display while also returning false
		// This dual approach enables both conditional logic in handlers and
		// user feedback through template error display mechanisms
		f.Errors.Add(field, "This field cannot be blank")
		return false
	}

	// Field contains content - validation passed
	return true
}

// MinLength implements the Strategy pattern for field length validation.
// This validation strategy ensures that form fields meet minimum length requirements,
// which is essential for data quality, security (password strength), and business
// rules (minimum name lengths, etc.). It demonstrates parameterized validation
// where the validation rule behavior can be customized based on requirements.
//
// Length validation is particularly important for user-generated content because:
// 1. It prevents accidentally empty submissions that pass basic presence checks
// 2. It enforces business rules about data quality (e.g., meaningful names)
// 3. It provides security benefits for fields like passwords or usernames
// 4. It enables consistent data formatting and presentation requirements
//
// Design Pattern: Strategy - implements parameterized length validation algorithm
// Design Pattern: Template Method - follows consistent validation result pattern
// Parameters:
//
//	field: Name of the form field to validate for minimum length
//	length: Minimum number of characters required for valid content
//
// Returns: true if field meets length requirement, false if too short
func (f *Form) MinLength(field string, length int) bool {
	// Retrieve the submitted value for length analysis
	// Using consistent field access patterns across all validation methods
	// maintains predictable behavior and makes the validation framework reliable
	x := f.Get(field)

	// Compare actual field length against minimum requirement
	if len(x) < length {
		// Generate contextual error message with specific length requirement
		// Including the actual requirement in the error message helps users
		// understand exactly what they need to do to fix the validation failure
		f.Errors.Add(field, fmt.Sprintf("This field must be at least %d characters long", length))
		return false
	}

	// Field meets minimum length requirement - validation passed
	return true

	// Note: This validation method checks string length in bytes, not Unicode code points
	// For international applications requiring Unicode-aware length checking,
	// consider using utf8.RuneCountInString(x) instead of len(x) for accurate
	// character counting across different languages and character sets
}

// IsEmail implements the Strategy pattern for email format validation.
// This validation strategy ensures that email fields contain syntactically valid
// email addresses according to standard email format specifications. It demonstrates
// how external validation libraries can be integrated into custom validation
// frameworks while maintaining consistent interfaces and error handling patterns.
//
// Email validation is critical for web applications because:
// 1. Invalid email addresses break communication workflows (notifications, password reset)
// 2. Email format errors are common user mistakes that need clear feedback
// 3. Email validation prevents data quality issues in customer databases
// 4. Proper email validation improves deliverability and reduces bounce rates
//
// Design Pattern: Strategy - implements email format validation algorithm
// Design Pattern: Adapter - integrates external govalidator library with internal validation framework
// Design Pattern: Library Integration - demonstrates clean integration of third-party validation
// Parameters:
//
//	field: Name of the form field containing email address to validate
func (f *Form) IsEmail(field string) {
	// Validate email format using specialized external library
	// The govalidator library implements comprehensive email format checking
	// according to RFC specifications, handling complex cases like internationalized
	// domains, quoted strings, and other edge cases that would be difficult to
	// implement correctly with simple regular expressions
	if !govalidator.IsEmail(f.Get(field)) {
		// Record clear, actionable error message for invalid email format
		// The error message focuses on the user action needed (provide valid email)
		// rather than technical details about email format specifications
		f.Errors.Add(field, "Invalid email address")
	}

	// Note: This validation checks format only, not deliverability
	// For production applications requiring email deliverability verification,
	// consider additional validation steps such as:
	// 1. DNS MX record checking for domain validity
	// 2. SMTP connection testing for mailbox existence
	// 3. Integration with email verification services
	// 4. Confirmation email workflows for address verification
	//
	// However, these advanced validations should typically be performed
	// asynchronously to avoid blocking form submission workflows
}
