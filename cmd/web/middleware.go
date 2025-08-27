package main

import (
	"net/http"

	"github.com/justinas/nosurf"
)

// NoSurf implements the Decorator pattern for CSRF (Cross-Site Request Forgery) protection.
// This middleware function wraps HTTP handlers with security protection that validates
// form submissions and AJAX requests to prevent malicious cross-site attacks. It demonstrates
// how security concerns can be handled transparently through the middleware chain without
// requiring modifications to individual handler functions.
//
// The Decorator pattern is particularly powerful here because it allows us to add security
// behavior to existing handlers without changing their code. This separation of concerns
// means that handlers can focus on business logic while middleware handles cross-cutting
// concerns like security, logging, and session management.
//
// Design Pattern: Decorator - adds security behavior to handlers without modifying them
// Design Pattern: Chain of Responsibility - part of the middleware processing pipeline
// Security Pattern: CSRF Protection - validates request authenticity using tokens
// Parameters:
//
//	next: The HTTP handler to be wrapped with CSRF protection
//
// Returns: A new handler that includes CSRF validation before calling the original handler
func NoSurf(next http.Handler) http.Handler {
	// Create the CSRF protection wrapper using the nosurf third-party library
	// This demonstrates how Go's middleware pattern integrates cleanly with external
	// security libraries while maintaining the standard http.Handler interface
	csrfHandler := nosurf.New(next)

	// Configure the CSRF token cookie with security-appropriate settings
	// These cookie configuration options balance security requirements with usability
	// across different deployment environments and browser configurations
	csrfHandler.SetBaseCookie(http.Cookie{
		// HttpOnly prevents client-side JavaScript from accessing the CSRF token
		// This is a crucial XSS (Cross-Site Scripting) protection mechanism that ensures
		// malicious scripts cannot steal CSRF tokens even if they execute on the page
		HttpOnly: true,

		// Path restricts the cookie to the entire application scope
		// Setting path to "/" means the CSRF token will be sent with requests to any
		// URL within the application, enabling protection across all protected endpoints
		Path: "/",

		// Secure flag controls whether cookies are sent only over HTTPS connections
		// In production (app.InProduction = true), this enforces encrypted transport
		// In development, it's disabled to allow HTTP testing without SSL certificates
		Secure: app.InProduction,

		// SameSite provides additional CSRF protection through browser-level controls
		// SameSiteLaxMode allows the cookie to be sent with top-level navigation (links)
		// but blocks it from being sent with cross-site POST requests, providing
		// a defense-in-depth security approach alongside the CSRF token validation
		SameSite: http.SameSiteLaxMode,
	})

	// Return the decorated handler that will process requests through CSRF validation
	// The returned handler maintains the same interface as the original, demonstrating
	// how the Decorator pattern preserves the interface contract while adding behavior
	return csrfHandler
}

// SessionLoad implements the Decorator pattern for HTTP session management.
// This middleware ensures that user session data is loaded from storage before
// request processing begins and is saved back to storage after the request completes.
// It demonstrates how stateful behavior can be added to the inherently stateless
// HTTP protocol through middleware that manages the session lifecycle transparently.
//
// Session management is a classic example of cross-cutting concerns in web applications.
// Rather than requiring every handler to manually load and save session data, this
// middleware handles the session lifecycle automatically, allowing handlers to simply
// read and write session values through the application's session manager.
//
// Design Pattern: Decorator - adds session management to handlers transparently
// Design Pattern: Chain of Responsibility - integrates with the middleware processing chain
// Design Pattern: Proxy - acts as a proxy for session storage operations
// Parameters:
//
//	next: The HTTP handler that will process the request with session data available
//
// Returns: A new handler that manages session loading and saving around the original handler
func SessionLoad(next http.Handler) http.Handler {
	// Use the session manager's LoadAndSave method to create the session-aware wrapper
	// This method implements the Proxy pattern by intercepting HTTP requests and responses
	// to handle session data persistence transparently to the wrapped handler
	//
	// The LoadAndSave wrapper performs the following sequence:
	// 1. Before request processing: Load existing session data from storage (cookies/database)
	// 2. During request processing: Make session data available via request context
	// 3. After request processing: Save any session changes back to storage
	//
	// This lifecycle management ensures that session state is consistent and persistent
	// across HTTP requests while maintaining the stateless nature of individual handlers
	return session.LoadAndSave(next)
}
