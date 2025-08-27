package helpers

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/bensabler/milos-residence/internal/config"
)

// app implements the Service Locator pattern for accessing application-wide configuration.
// This package-level variable provides helper functions with access to shared services
// including loggers for error reporting, session managers for user state, and configuration
// flags that control helper behavior across different deployment environments. It demonstrates
// how utility packages can access cross-cutting concerns while maintaining clean interfaces.
//
// The Service Locator pattern is appropriate here because helper functions represent shared
// utilities used throughout the application that benefit from consistent access to logging
// and configuration services. This approach eliminates the need to pass configuration
// parameters to every helper function call while ensuring consistent behavior across
// all error handling and utility operations.
//
// Design Pattern: Service Locator - provides access to shared application services
// Design Pattern: Dependency Injection - configuration injected during initialization
var app *config.AppConfig

// NewHelpers implements the Dependency Injection pattern for helper package initialization.
// This function connects the helpers package to the shared application configuration,
// enabling helper functions to access logging, session management, and other cross-cutting
// services. It demonstrates how utility packages can be initialized with the dependencies
// they need while maintaining clean, simple interfaces for the functions they provide.
//
// The dependency injection approach provides several architectural benefits:
// 1. **Explicit Dependencies**: Helper functions clearly declare their service requirements
// 2. **Testability**: Different configurations can be injected for testing scenarios
// 3. **Consistency**: All helper operations use the same logging and configuration services
// 4. **Flexibility**: Helper behavior can be modified through configuration without code changes
// 5. **Maintainability**: Configuration changes propagate automatically to all helper operations
//
// Design Pattern: Dependency Injection - receives external dependencies rather than creating them
// Design Pattern: Initialization - sets up package-level services for subsequent operations
// Parameters:
//
//	a: Application configuration providing access to loggers and shared services
func NewHelpers(a *config.AppConfig) {
	// Store application configuration reference for use by helper functions
	// This enables all helper functions to access shared services like loggers
	// without requiring configuration parameters in every function call
	app = a
}

// ClientError implements the Error Response pattern for handling client-side HTTP errors.
// This function provides standardized handling of HTTP errors that result from client
// mistakes or invalid requests (4xx status codes), ensuring consistent error responses
// and appropriate logging for operational monitoring. It demonstrates how web applications
// can differentiate between client errors and server errors while providing appropriate
// feedback and maintaining security best practices.
//
// Client errors represent problems with the request itself rather than server failures:
// 1. **Bad Request (400)**: Malformed request syntax or invalid parameters
// 2. **Unauthorized (401)**: Missing or invalid authentication credentials
// 3. **Forbidden (403)**: Valid request but insufficient permissions
// 4. **Not Found (404)**: Requested resource doesn't exist or isn't accessible
// 5. **Method Not Allowed (405)**: HTTP method not supported for the resource
//
// The function provides operational logging while avoiding detailed error exposure
// that could provide information useful to attackers attempting to probe system vulnerabilities.
//
// Design Pattern: Error Response - standardized HTTP error handling and response generation
// Design Pattern: Security by Obscurity - limits error detail exposure in client responses
// Design Pattern: Operational Monitoring - provides logging for system health tracking
// Parameters:
//
//	w: HTTP response writer for sending error response to client
//	status: HTTP status code indicating the type of client error (4xx range)
func ClientError(w http.ResponseWriter, status int) {
	// Log client error for operational monitoring and traffic analysis
	// This provides visibility into client behavior patterns, potential attack attempts,
	// and system usage issues that might require attention or configuration changes
	app.InfoLog.Println("Client error with status of", status)

	// Send standardized HTTP error response using status code's canonical text
	// http.StatusText provides the standard description (e.g., "Bad Request" for 400)
	// This approach ensures consistent, predictable error messages while avoiding
	// custom error text that might inadvertently expose system information
	http.Error(w, http.StatusText(status), status)

	// Note: Consider implementing request ID tracking in production systems
	// Request IDs enable correlation between client-visible errors and server logs,
	// facilitating customer support and system troubleshooting without exposing
	// sensitive system details in client-facing error messages
}

// ServerError implements the Error Response pattern for handling server-side HTTP errors.
// This function provides comprehensive error handling for internal system failures (5xx status codes),
// including detailed logging with stack traces for developer diagnosis while presenting
// generic error messages to clients for security. It demonstrates how production web applications
// balance developer debugging needs with security requirements and user experience considerations.
//
// Server errors indicate problems within the application or infrastructure rather than client mistakes:
// 1. **Internal Server Error (500)**: Unhandled application errors or system failures
// 2. **Service Unavailable (503)**: Temporary system overload or maintenance conditions
// 3. **Gateway Timeout (504)**: Upstream service failures or network connectivity issues
// 4. **Insufficient Storage (507)**: Disk space or storage quota exhaustion
//
// The comprehensive error logging enables rapid problem diagnosis and resolution while
// protecting users from technical details that could be confusing or potentially exploitable.
//
// Design Pattern: Error Response - comprehensive server error handling with security considerations
// Design Pattern: Error Logging - detailed diagnostic information for developer troubleshooting
// Design Pattern: Information Hiding - protects sensitive system details from client exposure
// Design Pattern: Stack Trace Capture - preserves execution context for effective debugging
// Parameters:
//
//	w: HTTP response writer for sending error response to client
//	err: Original error containing details needed for debugging and system diagnosis
func ServerError(w http.ResponseWriter, err error) {
	// Capture complete error context including stack trace for comprehensive debugging
	// runtime/debug.Stack() provides the call stack at the point of error occurrence,
	// enabling developers to understand exactly where and how the error originated
	// within the application's execution flow
	trace := fmt.Errorf("%s\n%s", err.Error(), debug.Stack())

	// Log detailed error information to centralized error logging system
	// This creates a permanent record of the error with full context for analysis,
	// troubleshooting, and system improvement without exposing details to users
	app.ErrorLog.Println(trace)

	// Send generic HTTP 500 response to client without exposing internal details
	// http.StatusText(http.StatusInternalServerError) returns "Internal Server Error"
	// This standard message provides appropriate feedback while protecting system internals
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

	// Security and operational considerations for production deployments:
	// 1. Ensure verbose stack traces are never included in client responses
	// 2. Consider implementing error correlation IDs for customer support
	// 3. Integrate with monitoring systems for automated alerting on error patterns
	// 4. Review error logs regularly to identify recurring issues requiring attention
	// 5. Implement error rate limiting to prevent log flooding during system issues
}
