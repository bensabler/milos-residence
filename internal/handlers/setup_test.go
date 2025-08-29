package handlers

import (
	"encoding/gob"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/bensabler/milos-residence/internal/config"
	"github.com/bensabler/milos-residence/internal/models"
	"github.com/bensabler/milos-residence/internal/render"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/justinas/nosurf"
)

// functions implements the Template Function Registry pattern for testing template functionality.
// This map stores custom functions that templates can invoke during test execution, providing
// the same template enhancement capabilities used in production while supporting isolated
// testing scenarios. It demonstrates how testing environments can replicate production
// template functionality without requiring complex production infrastructure setup.
//
// Template function testing ensures that:
// 1. **Template Compatibility**: Tests use the same template functions as production
// 2. **Function Isolation**: Test functions don't interfere with production template state
// 3. **Extensibility Testing**: New template functions can be tested before production deployment
// 4. **Error Handling**: Template function failures are properly handled during testing
// 5. **Performance Validation**: Template function performance can be measured in test scenarios
//
// The empty function map represents a clean slate for testing, allowing tests to add specific
// template functions as needed for particular test scenarios without carrying forward any
// production-specific template functions that might interfere with test predictability.
//
// Design Pattern: Function Registry - centralized repository of template helper functions for testing
// Design Pattern: Test Isolation - separate function registry prevents test interference
var functions = template.FuncMap{}

// app implements the Test Configuration pattern for handler testing environment setup.
// This package-level variable provides test-specific application configuration that enables
// comprehensive handler testing without requiring production infrastructure or external
// services. It demonstrates how testing environments can replicate production behavior
// while maintaining test isolation, predictability, and performance.
//
// Test configuration management provides several testing advantages:
// 1. **Environment Isolation**: Tests run with controlled configuration independent of production
// 2. **Service Mocking**: External dependencies can be replaced with test doubles
// 3. **Behavior Control**: Test-specific settings enable different testing scenarios
// 4. **Performance Optimization**: Test configuration optimized for speed rather than production requirements
// 5. **Error Simulation**: Configuration can be modified to simulate various error conditions
//
// The test configuration includes all components needed for comprehensive handler testing
// including logging, session management, template rendering, and database access through
// test doubles rather than production database connections.
//
// Design Pattern: Test Configuration - specialized configuration for testing scenarios
// Design Pattern: Service Locator - provides access to test-specific services and dependencies
var app config.AppConfig

// session implements the Test Session Management pattern for session-dependent handler testing.
// This session manager provides the same session functionality used in production while
// operating in a controlled testing environment that enables predictable session behavior,
// isolated test execution, and comprehensive testing of session-dependent application features.
//
// Test session management is crucial for web application testing because:
// 1. **Workflow Testing**: Multi-step user workflows require session state persistence across requests
// 2. **Authentication Testing**: Login/logout functionality depends on session state management
// 3. **Shopping Cart Testing**: E-commerce workflows store temporary state in session storage
// 4. **Flash Message Testing**: User feedback messages use session storage for display timing
// 5. **Security Testing**: Session security features must be validated under controlled conditions
//
// The test session manager uses the same underlying session library as production but with
// test-specific configuration that prioritizes predictability and test isolation over
// production performance and security requirements.
//
// Design Pattern: Test Session Management - controlled session functionality for testing
// Design Pattern: State Management Testing - validates stateful behavior in stateless HTTP
var session *scs.SessionManager

// pathToTemplates defines the template directory location for handler testing scenarios.
// This constant enables template-dependent handlers to access HTML templates during test
// execution, supporting comprehensive integration testing that includes template rendering,
// data binding, and complete response generation. It demonstrates how testing environments
// can access production assets while maintaining test isolation and predictable execution.
//
// Template path configuration for testing provides several benefits:
// 1. **Integration Testing**: Handlers can render actual templates during testing
// 2. **Template Validation**: Template syntax and data binding can be verified through handler tests
// 3. **Response Testing**: Complete HTML responses can be validated for content and structure
// 4. **Error Detection**: Template errors are discovered during test execution rather than production
// 5. **Performance Testing**: Template rendering performance can be measured in test scenarios
//
// The relative path approach supports both development and testing environments while
// maintaining consistent template organization and discovery patterns across different
// execution contexts.
//
// Design Pattern: Path Configuration - standardized template location for testing scenarios
// Design Pattern: Resource Access - enables test access to production template assets
var pathToTemplates = "./../../templates"

// TestMain implements the Test Suite Initialization pattern for comprehensive handler testing setup.
// This function serves as the centralized initialization point for all handler tests, configuring
// the complete testing environment including session management, template rendering, logging,
// and dependency injection. It demonstrates how complex testing environments can be established
// once per test suite rather than repeatedly for individual tests, improving test performance
// and ensuring consistent test conditions across all test scenarios.
//
// Test suite initialization provides several architectural advantages:
// 1. **Performance Optimization**: Expensive setup operations performed once rather than per test
// 2. **Environment Consistency**: All tests execute with identical configuration and dependencies
// 3. **Resource Management**: Test resources allocated and cleaned up systematically
// 4. **Isolation Assurance**: Test environment separated from production configuration
// 5. **Integration Support**: Complex testing scenarios enabled through comprehensive setup
//
// The initialization sequence replicates production application startup but with test-specific
// configuration that prioritizes predictability, performance, and isolation over production
// security and scalability requirements.
//
// Design Pattern: Test Suite Initialization - centralized setup for entire test package
// Design Pattern: Environment Configuration - test-specific environment setup and management
// Design Pattern: Dependency Injection - test dependencies injected during initialization
func TestMain(m *testing.M) {
	// Register data types for session storage during testing scenarios
	// Session serialization requires type registration using Go's gob package,
	// enabling complex objects like reservations and user data to be stored
	// in session state and retrieved during subsequent test requests
	//
	// Type registration is critical for session-based testing because:
	// - Session data must be serialized for storage between HTTP requests
	// - Complex business objects need explicit registration for proper serialization
	// - Test scenarios often store realistic business data in session state
	// - Missing type registration causes runtime panics during session access
	gob.Register(models.Reservation{})

	// Configure testing environment flags for test-appropriate behavior
	// InProduction = false enables development-friendly settings including:
	// - HTTP cookies (no HTTPS requirement) for easier test client usage
	// - Detailed error messages for better test debugging capabilities
	// - Relaxed security settings appropriate for isolated test environments
	// - Performance optimizations focused on test speed rather than production scale
	app.InProduction = false

	// Initialize INFO logging for test execution monitoring and debugging
	// Test logging provides visibility into handler execution, database operations,
	// and business logic processing during test scenarios, enabling effective
	// debugging when tests fail or behave unexpectedly
	infoLog := log.New(os.Stdout, "INFO:\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	// Initialize ERROR logging with enhanced context for test debugging
	// Error logging with file and line information (log.Lshortfile) enables
	// rapid identification of error sources during test execution, supporting
	// efficient debugging and problem resolution when tests encounter issues
	errorLog := log.New(os.Stdout, "ERROR:\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	// Initialize session management for test scenarios
	// Session configuration appropriate for testing includes shorter lifetimes,
	// relaxed security settings, and behavior optimized for test predictability
	// rather than production security and performance requirements
	session = scs.New()
	session.Lifetime = 24 * time.Hour              // Generous lifetime prevents session expiration during tests
	session.Cookie.Persist = true                  // Persistent cookies survive test scenario transitions
	session.Cookie.SameSite = http.SameSiteLaxMode // Relaxed SameSite for cross-origin test scenarios
	session.Cookie.Secure = app.InProduction       // HTTP-only cookies acceptable for test environments

	// Store session manager in application configuration for handler access
	// Handlers require session access through application configuration,
	// maintaining the same dependency injection patterns used in production
	app.Session = session

	mailChan := make(chan models.MailData)
	app.MailChan = mailChan
	defer close(mailChan)

	listenForMail()

	// Create template cache for handler testing with template rendering
	// Template cache enables handlers to render complete HTML responses during
	// testing, supporting comprehensive integration testing that validates
	// both handler logic and template rendering behavior together
	tc, err := CreateTestTemplateCache()
	if err != nil {
		// Template cache creation failed - this prevents template-dependent tests
		// Fatal error terminates test execution because handlers cannot function
		// without template rendering capabilities for complete response generation
		log.Fatal("cannot create template cache: ", err)
	}
	app.TemplateCache = tc

	// Configure template caching behavior for test performance optimization
	// UseCache = true enables template reuse across test scenarios, providing
	// faster test execution while still validating template functionality
	// Test environments benefit from template caching for performance reasons
	app.UseCache = true

	// Initialize handler dependencies using test-specific configuration
	// Dependency injection setup replicates production initialization patterns
	// but uses test doubles and controlled configuration for predictable testing
	repo := NewTestRepo(&app) // Create repository with test database doubles
	NewHandlers(repo)         // Initialize handlers with test repository
	render.NewRenderer(&app)  // Configure template renderer with test settings

	// Execute all tests in the package and terminate with appropriate exit code
	// os.Exit ensures that test process terminates with correct exit status
	// for integration with build systems and continuous integration pipelines
	os.Exit(m.Run())
}

func listenForMail() {
	go func() {
		for {
			_ = <-app.MailChan
		}
	}()
}

// getRoutes implements the Test Router Factory pattern for handler integration testing.
// This function creates a complete HTTP router configuration specifically designed for
// testing scenarios, including all middleware, route mappings, and static file handling
// needed for comprehensive handler testing. It demonstrates how testing environments can
// replicate production routing behavior while maintaining test-specific optimizations
// and controlled execution conditions.
//
// Test router configuration provides several testing advantages:
// 1. **Integration Testing**: Complete request processing pipeline from routing to response
// 2. **Middleware Testing**: Validates middleware integration with handlers under test conditions
// 3. **Route Coverage**: Systematic testing of all application routes with consistent configuration
// 4. **Performance Testing**: Router performance can be measured under controlled test conditions
// 5. **Error Handling**: Route-level error handling validated through integration testing
//
// The router configuration matches production setup while using test-specific middleware
// configuration that prioritizes test reliability and debugging support over production
// security and performance optimizations.
//
// Design Pattern: Test Router Factory - creates complete router configuration for testing
// Design Pattern: Integration Testing Setup - enables end-to-end request processing testing
// Design Pattern: Middleware Integration - validates middleware behavior within complete routing system
func getRoutes() http.Handler {
	// Create chi router instance for test request routing
	// Chi router provides the same routing capabilities used in production,
	// enabling realistic testing of URL pattern matching, parameter extraction,
	// and HTTP method routing under controlled test conditions
	mux := chi.NewRouter()

	// Apply essential middleware for handler integration testing
	// Middleware configuration includes components needed for comprehensive
	// handler testing while omitting or modifying middleware that might
	// interfere with test execution or require external dependencies

	// Panic recovery middleware prevents test failures from crashing test runner
	// Recoverer converts handler panics into HTTP 500 responses, enabling tests
	// to validate error handling behavior rather than terminating test execution
	mux.Use(middleware.Recoverer)

	// Session management middleware enables session-dependent handler testing
	// SessionLoad provides the same session functionality used in production
	// but with test-specific configuration optimized for test predictability
	mux.Use(SessionLoad)

	// Note: CSRF middleware (NoSurf) commented out for testing simplicity
	// CSRF protection adds complexity to test request preparation and may not
	// be necessary for basic handler functionality testing. Can be enabled
	// for tests that specifically validate CSRF protection behavior

	// Configure application routes matching production route configuration
	// Route mapping enables systematic testing of all application endpoints
	// with consistent middleware application and handler integration

	// Core application pages - essential user-facing functionality
	mux.Get("/", Repo.Home)         // Landing page handler
	mux.Get("/about", Repo.About)   // Information page handler
	mux.Get("/photos", Repo.Photos) // Gallery page handler

	// Themed accommodation detail pages - booking workflow support
	mux.Get("/golden-haybeam-loft", Repo.GoldenHaybeamLoft)   // Room detail handler
	mux.Get("/window-perch-theater", Repo.WindowPerchTheater) // Room detail handler
	mux.Get("/laundry-basket-nook", Repo.LaundryBasketNook)   // Room detail handler

	// Booking system functionality - core business logic handlers
	mux.Get("/search-availability", Repo.Availability)           // Availability search form
	mux.Post("/search-availability", Repo.PostAvailability)      // Search processing handler
	mux.Post("/search-availability-json", Repo.AvailabilityJSON) // AJAX availability API

	// Customer communication - contact and support functionality
	mux.Get("/contact", Repo.Contact) // Contact information handler

	// Reservation workflow - booking process handlers
	mux.Get("/make-reservation", Repo.MakeReservation)       // Reservation form handler
	mux.Post("/make-reservation", Repo.PostReservation)      // Reservation processing handler
	mux.Get("/reservation-summary", Repo.ReservationSummary) // Booking confirmation handler

	// Static file serving for complete application testing
	// File server enables testing of complete application functionality including
	// CSS, JavaScript, and image assets needed for realistic testing scenarios
	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	// Return configured router ready for handler integration testing
	// Router includes complete middleware chain, route configuration, and
	// static file handling needed for comprehensive handler testing scenarios
	return mux
}

// NoSurf implements the Test CSRF Protection pattern for security-aware handler testing.
// This middleware provides the same CSRF protection functionality used in production
// but with test-friendly configuration that enables comprehensive testing of form
// submission security while maintaining test simplicity and execution speed.
//
// CSRF protection testing is important because:
// 1. **Security Validation**: Form submission security must work correctly in all environments
// 2. **Integration Testing**: CSRF tokens must integrate properly with form processing handlers
// 3. **Error Testing**: CSRF validation failures should be handled gracefully with appropriate user feedback
// 4. **Token Generation**: CSRF token generation and validation must work consistently across requests
// 5. **Cookie Management**: CSRF cookies must be configured correctly for security and usability
//
// Test CSRF configuration balances security validation with test execution simplicity,
// enabling comprehensive security testing without requiring complex test setup procedures.
//
// Design Pattern: Test Security Middleware - CSRF protection configured for testing scenarios
// Design Pattern: Security Testing - validates security features under controlled conditions
func NoSurf(next http.Handler) http.Handler {
	// Create CSRF protection wrapper using nosurf library
	// Same CSRF implementation used in production ensures that test validation
	// accurately reflects production security behavior and identifies security
	// issues that might affect real user interactions
	csrfHandler := nosurf.New(next)

	// Configure CSRF cookie with test-appropriate security settings
	// Cookie configuration balances security validation with test simplicity,
	// enabling CSRF testing without requiring complex HTTPS setup or certificate management
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,                 // Prevent JavaScript access for XSS protection testing
		Path:     "/",                  // Cookie scope covers entire application for complete testing
		Secure:   app.InProduction,     // HTTP cookies acceptable for test environments
		SameSite: http.SameSiteLaxMode, // Relaxed SameSite policy for test cross-origin scenarios
	})

	// Return CSRF-protected handler ready for security testing
	// Protected handler validates CSRF tokens on state-changing requests while
	// maintaining the http.Handler interface for integration with test infrastructure
	return csrfHandler
}

// SessionLoad implements the Test Session Middleware pattern for session-aware handler testing.
// This middleware provides the same session management functionality used in production
// but operating within the test environment's session manager configuration. It enables
// comprehensive testing of session-dependent handlers while maintaining test isolation
// and predictable session behavior across test scenarios.
//
// Session middleware testing validates:
// 1. **Session Lifecycle**: Session loading and saving works correctly during request processing
// 2. **Context Integration**: Session data is properly attached to request context for handler access
// 3. **Error Handling**: Session failures are handled gracefully without breaking request processing
// 4. **State Persistence**: Session state persists correctly across multiple test requests
// 5. **Memory Management**: Session resources are properly managed during test execution
//
// The test session middleware uses the same session loading and saving patterns as production
// while operating with test-specific session configuration optimized for test predictability.
//
// Design Pattern: Test Session Middleware - session management configured for testing scenarios
// Design Pattern: State Management Testing - validates stateful behavior in test environments
func SessionLoad(next http.Handler) http.Handler {
	// Use session manager's LoadAndSave wrapper for complete session lifecycle management
	// This provides the same session handling used in production, including:
	// - Session loading from storage before request processing
	// - Session data availability through request context during handler execution
	// - Session saving to storage after request processing completes
	// - Error handling for session failures that might occur during testing
	return session.LoadAndSave(next)
}

// CreateTestTemplateCache implements the Test Template Compilation pattern for handler template testing.
// This function creates a template cache specifically configured for testing scenarios, enabling
// handlers to render complete HTML responses during test execution while maintaining test
// performance and isolation. It demonstrates how testing environments can replicate production
// template functionality while optimizing for test-specific requirements.
//
// Test template compilation provides several testing advantages:
// 1. **Integration Testing**: Handlers can render actual templates for complete response validation
// 2. **Template Validation**: Template syntax and compilation errors discovered during test setup
// 3. **Performance Testing**: Template rendering performance measured under controlled conditions
// 4. **Error Detection**: Template data binding issues identified through handler testing
// 5. **Response Testing**: Complete HTML responses validated for content, structure, and correctness
//
// The template compilation process matches production template handling while using test-specific
// paths and configuration that support test isolation and predictable template behavior.
//
// Design Pattern: Test Template Compilation - template cache creation optimized for testing
// Design Pattern: Resource Management - efficient template compilation and caching for tests
// Returns: Template cache ready for use in handler testing scenarios, or error if compilation fails
func CreateTestTemplateCache() (map[string]*template.Template, error) {
	// Initialize empty template cache for test template storage
	// Cache provides the same template storage and lookup functionality used in
	// production while operating with test-specific template discovery and compilation
	myCache := map[string]*template.Template{}

	// Discover page templates using test-specific template path configuration
	// Template discovery finds all page templates available for handler testing,
	// enabling comprehensive testing of template-dependent handlers across the application
	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.tmpl", pathToTemplates))
	if err != nil {
		// Template discovery failed - return error to prevent test execution with incomplete templates
		// Template discovery failures indicate file system access issues or incorrect path configuration
		return myCache, err
	}

	// Compile and cache each discovered page template for handler testing
	// Template compilation process matches production template processing while
	// using test-specific function registration and layout template association
	for _, page := range pages {
		// Extract template name for cache key generation
		// Template name serves as cache lookup key during handler test execution
		name := filepath.Base(page)

		// Create template instance with function registry for advanced template testing
		// Function registration enables testing of templates that use custom template
		// functions for data formatting, calculations, or presentation logic
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			// Page template compilation failed - return error with context for debugging
			return myCache, err
		}

		// Associate layout templates with page templates for complete rendering capability
		// Layout template integration enables testing of complete page rendering including
		// shared navigation, headers, footers, and other common page elements
		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))
		if err != nil {
			// Layout template discovery failed - return error to prevent incomplete template sets
			return myCache, err
		}

		// Integrate layout templates into page template compilation
		if len(matches) > 0 {
			// ParseGlob associates all layout templates with current page template
			// This enables complete page rendering with shared layout elements during testing
			if ts, err = ts.ParseGlob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates)); err != nil {
				// Layout template integration failed - return error with compilation context
				return myCache, err
			}
		}

		// Store completed template set in cache for handler testing access
		// Cached template includes page content and associated layouts, ready for
		// complete HTML response generation during handler test execution
		myCache[name] = ts
	}

	// Return completed template cache ready for comprehensive handler testing
	// Template cache enables handlers to render complete HTML responses during
	// test execution, supporting integration testing of template-dependent functionality
	return myCache, nil
}
