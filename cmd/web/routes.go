// Package main implements the web server entry point and request routing infrastructure.
// This file demonstrates how modern Go web applications organize HTTP request routing
// and apply cross-cutting concerns through middleware chains. It serves as the central
// configuration point where URLs are mapped to handlers and security policies are applied.
package main

import (
	"net/http"

	"github.com/bensabler/milos-residence/internal/config"
	"github.com/bensabler/milos-residence/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// routes implements the Front Controller pattern for centralized request routing.
// This function creates the main HTTP request dispatcher that determines which handler
// should process each incoming request based on URL patterns and HTTP methods. It
// demonstrates how modern web applications organize routing logic and apply security
// middleware consistently across all endpoints.
//
// The Front Controller pattern provides several architectural benefits:
// 1. Centralized request processing logic and security policy enforcement
// 2. Consistent middleware application across all routes without duplication
// 3. Clear mapping between URLs and business functionality for maintainability
// 4. Single point of configuration for cross-cutting concerns like CORS, logging, etc.
//
// Design Pattern: Front Controller - single entry point for all HTTP requests
// Design Pattern: Chain of Responsibility - middleware processing pipeline
// Design Pattern: Factory Method - creates configured router with all dependencies
// Parameters:
//
//	app: Application configuration containing shared services and settings
//
// Returns: Fully configured HTTP handler ready for use by the web server
func routes(app *config.AppConfig) http.Handler {
	// Initialize the chi router implementing the Router pattern for URL matching
	// Chi is chosen for its lightweight nature, composability, and compatibility with
	// the standard Go HTTP ecosystem. It provides advanced routing features like
	// URL parameters, middleware chains, and sub-routing while maintaining simplicity
	mux := chi.NewRouter()

	// Apply middleware stack using the Chain of Responsibility pattern
	// The order of middleware application is critical - each request flows through
	// these middleware functions in the specified sequence before reaching handlers.
	// This demonstrates how cross-cutting concerns are handled consistently across
	// the entire application without code duplication in individual handlers.

	// 1. Panic Recovery Middleware - implements the Circuit Breaker pattern
	// Recoverer catches any panics that occur during request processing and converts
	// them into proper HTTP 500 responses instead of crashing the entire server.
	// This is essential for production stability - a single malformed request or
	// programming error shouldn't bring down the entire application.
	mux.Use(middleware.Recoverer)

	// 2. CSRF Protection Middleware - implements the Security Filter pattern
	// NoSurf validates that form submissions and state-changing requests include
	// valid CSRF tokens to prevent cross-site request forgery attacks. This
	// middleware must run before handlers that process forms or AJAX requests.
	mux.Use(NoSurf)

	// 3. Session Management Middleware - implements the Session State pattern
	// SessionLoad ensures user session data is available to handlers via request
	// context and handles session persistence after request processing completes.
	// This must run after CSRF (which may use session storage) but before handlers.
	mux.Use(SessionLoad)

	// Define application routes using RESTful URL design principles
	// Each route maps a specific URL pattern and HTTP method to a handler function
	// that implements the Controller pattern for processing that type of request.
	// The URL design follows conventional web application patterns for usability.

	// Core Application Pages - implement the Page Controller pattern
	// These routes serve the primary user-facing pages of the application,
	// each handling a distinct area of functionality or content presentation.

	// Home page - serves as the application landing page and primary entry point
	// Implements the Front Page pattern for user onboarding and feature discovery
	mux.Get("/", handlers.Repo.Home)

	// About page - provides information about the business and establishes trust
	// Implements the Content Page pattern for static informational content
	mux.Get("/about", handlers.Repo.About)

	// Photo gallery - showcases visual content to support booking decisions
	// Implements the Gallery pattern for visual content presentation
	mux.Get("/photos", handlers.Repo.Photos)

	// Themed Room Detail Pages - implement the Product Detail pattern
	// Each route serves detailed information about a specific accommodation type,
	// supporting the customer decision-making process with rich content and booking integration.

	// Golden Haybeam Loft detail page - luxury accommodation showcase
	// Demonstrates themed content presentation with booking integration
	mux.Get("/golden-haybeam-loft", handlers.Repo.GoldenHaybeamLoft)

	// Window Perch Theater detail page - specialized accommodation features
	// Shows how different room types are presented with consistent UI patterns
	mux.Get("/window-perch-theater", handlers.Repo.WindowPerchTheater)

	// Laundry Basket Nook detail page - cozy accommodation emphasis
	// Completes the accommodation showcase with varied positioning strategies
	mux.Get("/laundry-basket-nook", handlers.Repo.LaundryBasketNook)

	// Availability Search System - implements the Search and Filter pattern
	// These routes handle the core booking workflow from availability checking
	// through room selection, demonstrating both traditional HTML forms and
	// modern AJAX interactions within the same application architecture.

	// Availability search form - displays date picker interface for availability checking
	// Implements the Search Interface pattern with progressive enhancement capabilities
	mux.Get("/search-availability", handlers.Repo.Availability)

	// Availability search processing - handles form submission and displays results
	// Implements the Search Results pattern with session-based result persistence
	mux.Post("/search-availability", handlers.Repo.PostAvailability)

	// JSON availability endpoint - provides AJAX API for dynamic availability checking
	// Implements the REST API pattern for single-page application functionality
	// This enables rich client-side interactions without full page refreshes
	mux.Post("/search-availability-json", handlers.Repo.AvailabilityJSON)

	// Room selection from search results - handles user choice from available options
	// Implements the Selection Handler pattern with URL parameter processing
	// The {id} parameter demonstrates RESTful resource identification in URLs
	mux.Get("/choose-room/{id}", handlers.Repo.ChooseRoom)

	// Direct booking from room pages - handles booking initiated from room detail pages
	// Implements the Quick Booking pattern with query parameter processing
	// This provides an alternative booking flow that bypasses the search interface
	mux.Get("/book-room", handlers.Repo.BookRoom)

	// Customer Communication - implements the Contact Interface pattern
	// Provides multiple channels for customer inquiries and support requests
	mux.Get("/contact", handlers.Repo.Contact)

	// Reservation Workflow - implements the Multi-Step Form pattern
	// These routes handle the complete booking process from form display through
	// confirmation, demonstrating session-based workflow management and the
	// Post-Redirect-Get pattern for form processing.

	// Reservation form display - shows booking form with pre-populated data from session
	// Implements the Form Display pattern with session-based state management
	mux.Get("/make-reservation", handlers.Repo.MakeReservation)

	// Reservation form processing - validates and saves booking information
	// Implements the Form Processing pattern with comprehensive validation and error handling
	// Uses POST method following HTTP semantics for state-changing operations
	mux.Post("/make-reservation", handlers.Repo.PostReservation)

	// Reservation confirmation - displays booking summary and next steps
	// Implements the Confirmation Page pattern with session cleanup and success messaging
	mux.Get("/reservation-summary", handlers.Repo.ReservationSummary)

	// Static Asset Serving - implements the Static Content pattern
	// This configuration serves CSS, JavaScript, images, and other static files
	// directly from the filesystem with appropriate caching headers for performance.
	// The http.FileServer implements efficient file serving with proper MIME types,
	// conditional requests (If-Modified-Since), and range request support.

	// Create file server for static assets from the ./static directory
	// FileServer implements optimized static content delivery with:
	// - Automatic MIME type detection based on file extensions
	// - HTTP caching headers for browser performance optimization
	// - Security headers to prevent directory traversal attacks
	fileServer := http.FileServer(http.Dir("./static/"))

	// Mount static file handler with URL prefix stripping
	// The /static/* pattern matches all URLs beginning with /static/
	// StripPrefix removes "/static" from the URL before file system lookup,
	// so /static/css/styles.css becomes a request for ./static/css/styles.css
	// This pattern allows flexible static asset organization while maintaining clean URLs
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	// Return the fully configured router ready for HTTP server use
	// The returned handler integrates all middleware, routes, and static asset serving
	// into a single http.Handler interface that the standard library HTTP server can use
	return mux
}
