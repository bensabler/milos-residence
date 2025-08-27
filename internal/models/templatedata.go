// Package models defines data structures shared between the business logic and presentation layers.
// This file specifically handles the View Model pattern implementation that bridges the gap
// between domain entities and template rendering requirements. It demonstrates how Go web
// applications can cleanly separate business data from presentation concerns while providing
// templates with exactly the data they need in the format they need it.
package models

import "github.com/bensabler/milos-residence/internal/forms"

// TemplateData implements the View Model pattern for template rendering data management.
// This struct serves as the standardized data container that carries information from
// HTTP handlers to HTML templates, providing a consistent interface for template rendering
// across the entire application. It demonstrates how web applications can organize template
// data to support complex user interfaces while maintaining clean separation between
// business logic and presentation logic.
//
// The View Model pattern is essential for maintainable web applications because it:
// 1. Provides a stable interface between controllers and views that doesn't change when business logic evolves
// 2. Enables templates to access exactly the data they need without tight coupling to domain entities
// 3. Supports complex user interface requirements like validation errors, flash messages, and dynamic content
// 4. Allows the same business data to be presented differently across different templates or contexts
// 5. Facilitates testing by providing a clear contract for what data templates expect to receive
//
// This pattern scales well from simple content pages to complex interactive forms because
// it provides flexible data organization without constraining how templates consume the data.
//
// Design Pattern: View Model - bridges business logic and presentation layer data requirements
// Design Pattern: Data Transfer Object - structured data container for template consumption
// Design Pattern: Facade - provides simplified interface to complex underlying data structures
type TemplateData struct {
	// StringMap provides key-value pairs for simple string data that templates consume directly.
	// This field handles the common case where templates need access to simple text values like
	// formatted dates, user-friendly labels, or computed display strings. Using a map enables
	// templates to access values by meaningful names rather than relying on struct field positions.
	//
	// Typical usage includes:
	//   - Formatted dates for display ("start_date" -> "January 15, 2024")
	//   - User-friendly status labels ("booking_status" -> "Confirmed")
	//   - Computed display values ("total_nights" -> "3 nights")
	//   - Localized text strings for internationalization support
	//   - URL fragments for dynamic link construction
	//
	// The map structure provides O(1) lookup performance and enables templates to use
	// descriptive key names that make template code more readable and maintainable
	StringMap map[string]string

	// IntMap provides key-value pairs for integer data that templates use for calculations or display.
	// This field handles numeric data that templates need for conditional logic, counting operations,
	// or mathematical calculations within template expressions. Separating integers from strings
	// maintains type safety and enables proper numeric operations in template logic.
	//
	// Typical usage includes:
	//   - Counts and quantities ("room_count" -> 5, "nights_booked" -> 3)
	//   - Status codes and enumeration values ("reservation_status" -> 1)
	//   - Pagination data ("current_page" -> 2, "total_pages" -> 10)
	//   - Configuration values ("max_guests" -> 4, "min_stay" -> 2)
	//   - Performance metrics ("response_time_ms" -> 150)
	IntMap map[string]int

	// FloatMap provides key-value pairs for floating-point data used in financial and measurement contexts.
	// This field handles decimal values that templates need for displaying prices, percentages, ratings,
	// or other precise numeric data. Using float32 provides sufficient precision for most web application
	// display purposes while minimizing memory usage and serialization overhead.
	//
	// Typical usage includes:
	//   - Pricing information ("room_rate" -> 149.99, "tax_amount" -> 12.75)
	//   - Ratings and scores ("average_rating" -> 4.7, "satisfaction_score" -> 8.5)
	//   - Percentages and ratios ("occupancy_rate" -> 0.85, "discount_percent" -> 15.0)
	//   - Measurements and quantities ("room_size_sqft" -> 425.5)
	//   - Financial calculations ("total_amount" -> 487.23, "deposit_required" -> 125.00)
	FloatMap map[string]float32

	// Data provides a flexible container for complex objects that don't fit into the typed maps above.
	// This field serves as the "escape hatch" for passing rich domain objects, collections, and
	// complex data structures to templates while maintaining the type safety and organization
	// provided by the more specific maps. It demonstrates Go's interface{} flexibility for
	// handling heterogeneous data requirements.
	//
	// Typical usage includes:
	//   - Domain entities ("reservation" -> models.Reservation, "user" -> models.User)
	//   - Collections and slices ("available_rooms" -> []models.Room, "recent_bookings" -> []models.Reservation)
	//   - Nested data structures ("booking_details" -> complex nested structs)
	//   - Third-party API responses ("weather_data" -> external service response)
	//   - File upload information ("uploaded_images" -> file metadata structures)
	//
	// While this provides maximum flexibility, use the typed maps above when possible
	// for better type safety and template clarity. Reserve Data for truly complex scenarios
	// where the structured approach of typed maps becomes limiting or overly verbose
	Data map[string]interface{}

	// CSRFToken provides Cross-Site Request Forgery protection for form submissions.
	// This field contains the unique token that must be included in all state-changing
	// HTTP requests (POST, PUT, DELETE) to verify that the request originated from the
	// application's own forms rather than from malicious third-party sites. It integrates
	// with the CSRF middleware to provide transparent security protection.
	//
	// The CSRF token is automatically generated per request and must be included in:
	//   - Hidden form fields: <input type="hidden" name="csrf_token" value="{{.CSRFToken}}">
	//   - AJAX request headers: X-CSRF-Token header with token value
	//   - URL parameters for non-form submissions: ?csrf_token={{.CSRFToken}}
	//
	// This field is automatically populated by the render.AddDefaultData function,
	// ensuring that all templates have access to the current request's CSRF token
	// without requiring explicit token management in individual handlers
	CSRFToken string

	// Flash provides temporary success messages for user feedback across HTTP redirects.
	// This field contains positive confirmation messages that inform users about successful
	// operations like "Reservation confirmed!" or "Profile updated successfully!". Flash
	// messages are stored in the user's session and automatically cleared after display,
	// preventing message duplication when users refresh pages or navigate backward.
	//
	// Flash messages implement the Post-Redirect-Get (PRG) pattern for web form processing:
	// 1. User submits form (POST request)
	// 2. Handler processes form and stores success message in session
	// 3. Handler redirects to confirmation page (Redirect response)
	// 4. Browser requests confirmation page (GET request)
	// 5. Template displays flash message and removes it from session
	//
	// This pattern prevents form resubmission problems and provides clean user experience
	// with clear feedback about operation success without cluttering the URL or browser history
	Flash string

	// Warning provides temporary caution messages for user awareness across HTTP redirects.
	// This field contains advisory messages that alert users to non-critical issues or
	// important information like "Your session will expire in 5 minutes" or "Some features
	// may not work properly with your browser settings". Warning messages use the same
	// session-based storage and automatic cleanup pattern as Flash messages.
	//
	// Warning messages serve different purposes than Flash or Error messages:
	//   - Flash: Positive confirmation of successful operations
	//   - Warning: Advisory information that doesn't require immediate action
	//   - Error: Critical problems that require user action to resolve
	//
	// This semantic separation enables templates to style and position different message
	// types appropriately, providing clear visual hierarchy that helps users understand
	// the relative importance and urgency of different types of feedback
	Warning string

	// Error provides temporary error messages for user feedback about operation failures.
	// This field contains critical error information that users must understand and act upon,
	// such as "Invalid email address" or "Credit card payment failed". Error messages follow
	// the same session-based storage pattern as Flash and Warning messages, ensuring they
	// persist across redirects but don't accumulate over multiple requests.
	//
	// Error messages are particularly important for user experience because they:
	//   - Explain why an operation failed in user-friendly language
	//   - Provide actionable guidance about what users need to do to resolve problems
	//   - Maintain context across the POST-Redirect-GET pattern used in form processing
	//   - Enable consistent error display across all application templates
	//
	// Error messages should be specific enough to help users fix problems but general enough
	// to avoid exposing sensitive system information that could be exploited by attackers.
	// Technical error details should be logged separately for developer diagnosis
	Error string

	// Form provides access to form validation state and error messages for interactive forms.
	// This field contains a Form instance from the internal/forms package that includes both
	// the submitted form data and any validation errors discovered during form processing.
	// It enables templates to repopulate form fields with user-submitted values and display
	// field-specific error messages inline with form inputs.
	//
	// The Form integration enables sophisticated form user experience patterns:
	//   - Pre-population: Templates can redisplay user input when validation fails
	//   - Inline errors: Field-specific error messages appear next to relevant inputs
	//   - Progressive enhancement: Basic HTML form validation enhanced with server-side validation
	//   - Accessibility: Screen readers can associate error messages with form fields
	//   - User efficiency: Users don't lose their input when validation fails
	//
	// Form is typically populated in two scenarios:
	// 1. GET requests: Empty form for initial display with no validation state
	// 2. POST requests with validation errors: Form with user data and error messages
	//
	// When validation succeeds, the handler typically redirects to a confirmation page
	// rather than re-rendering the form template, implementing the PRG pattern
	Form *forms.Form
}
