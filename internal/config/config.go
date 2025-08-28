package config

import (
	"html/template"
	"log"

	"github.com/alexedwards/scs/v2"
	"github.com/bensabler/milos-residence/internal/models"
)

// AppConfig implements the Configuration Object pattern for centralized application settings.
// This struct serves as the single source of truth for application-wide configuration,
// shared services, and behavioral flags. It demonstrates how Go applications can manage
// complex configuration and dependency injection without relying on global variables
// scattered throughout the codebase or heavyweight dependency injection frameworks.
//
// The Configuration Object pattern provides several architectural benefits:
// 1. Centralized configuration management - all settings accessible from one place
// 2. Explicit dependencies - components receive exactly what they need via injection
// 3. Environment-aware behavior - single flags control multiple behaviors across deployment stages
// 4. Testability - easy to create different configurations for testing vs production
// 5. Type safety - compile-time verification of configuration structure
//
// This approach scales well from simple applications to complex enterprise systems by
// providing a foundation for sophisticated configuration management without external
// dependencies or complex initialization sequences.
//
// Design Pattern: Configuration Object - centralized application settings and services
// Design Pattern: Service Locator - provides access to shared application services
// Design Pattern: Dependency Injection - enables explicit dependency management
type AppConfig struct {
	// UseCache controls template rendering strategy using the Strategy pattern.
	// This boolean flag determines whether the application uses cached templates
	// (true for production performance) or rebuilds templates on each request
	// (false for development rapid iteration). This single flag demonstrates how
	// configuration can control complex behavioral differences across environments.
	//
	// When true (production):
	//   - Templates parsed once at startup and cached in memory
	//   - Maximum rendering performance with minimal CPU usage
	//   - Template changes require application restart to take effect
	//
	// When false (development):
	//   - Templates parsed fresh on every request
	//   - Immediate visibility of template changes without restart
	//   - Higher CPU usage and slower response times (acceptable in development)
	UseCache bool

	// TemplateCache implements the Cache pattern for HTML template storage.
	// This map stores compiled Go templates keyed by their filename, providing
	// O(1) lookup performance for template rendering operations. The cache eliminates
	// expensive file I/O and template parsing operations during request processing,
	// which is critical for production performance under load.
	//
	// The key (string) represents the template filename (e.g., "home.page.tmpl")
	// The value (*template.Template) is the compiled template ready for execution
	// This design supports template inheritance, includes, and complex layouts
	// while maintaining fast access to any template in the system.
	TemplateCache map[string]*template.Template

	// InfoLog implements the Observer pattern for operational event logging.
	// This logger handles general application events like startup messages,
	// request processing milestones, and normal operational status. It provides
	// visibility into application health and behavior for system administrators
	// and DevOps teams monitoring production deployments.
	//
	// Typical usage includes:
	//   - Application startup and shutdown events
	//   - Database connection establishment
	//   - Request processing milestones
	//   - Configuration changes and feature toggles
	//   - Performance metrics and timing information
	InfoLog *log.Logger

	// ErrorLog implements the Observer pattern for error event logging.
	// This logger captures application errors, exceptions, and failure conditions
	// with enhanced context information including stack traces and request details.
	// The separation from InfoLog allows different log levels, destinations, and
	// processing pipelines for error conditions versus normal operations.
	//
	// Enhanced error logging includes:
	//   - Stack traces for debugging (via log.Lshortfile flag)
	//   - Request context and user session information
	//   - Database errors and connection failures
	//   - Template rendering failures and missing resources
	//   - External service integration failures
	ErrorLog *log.Logger

	// InProduction controls environment-specific behavior using the Strategy pattern.
	// This boolean flag serves as a master switch that influences multiple aspects
	// of application behavior, security settings, and performance optimizations.
	// It demonstrates how a single configuration value can coordinate complex
	// behavioral changes across different system components.
	//
	// When true (production environment):
	//   - HTTPS-only cookies for security (Secure flag set)
	//   - Template caching enabled for performance (UseCache typically true)
	//   - Detailed error messages hidden from users for security
	//   - Performance monitoring and metrics collection enabled
	//   - External service integrations use production endpoints
	//
	// When false (development/testing environment):
	//   - HTTP cookies allowed for local development convenience
	//   - Template caching disabled for rapid iteration (UseCache typically false)
	//   - Detailed error messages shown to developers for debugging
	//   - Debug logging and development tools enabled
	//   - External services may use sandbox/testing endpoints
	InProduction bool

	// Session implements the Session State pattern for user state management.
	// This session manager enables maintaining user state and data across the
	// stateless HTTP protocol through secure cookie-based session storage.
	// It provides the foundation for user authentication, shopping carts,
	// multi-step forms, flash messages, and any feature requiring state persistence.
	//
	// The session manager handles:
	//   - Secure session token generation and validation
	//   - Cookie-based session storage with configurable security settings
	//   - Session lifetime management and automatic expiration
	//   - Cross-site request forgery (CSRF) protection integration
	//   - Session data serialization and deserialization using gob encoding
	//
	// Session data can include:
	//   - User authentication status and identity
	//   - Shopping cart contents and temporary user preferences
	//   - Multi-step form data (like reservation workflows)
	//   - Flash messages for user feedback across redirects
	//   - User interface state and personalization settings
	Session  *scs.SessionManager
	MailChan chan models.MailData
}
