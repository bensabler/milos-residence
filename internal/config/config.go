// Package config centralizes application-wide configuration and shared resources.
// It exposes AppConfig, a single struct that is populated during startup and
// passed into packages that require access to global state (template cache,
// logging, session manager, mail channel, and prod/dev switches).
package config

import (
	"html/template"
	"log"

	"github.com/alexedwards/scs/v2"
	"github.com/bensabler/milos-residence/internal/models"
)

// AppConfig defines the process-wide application configuration and dependencies.
// Fields are intentionally concrete to keep wiring explicit and test-friendly.
//
// Typical initialization occurs in main() during bootstrap, after which a pointer
// to AppConfig is provided to packages that require these resources (e.g., render,
// handlers, helpers). Fields should be considered long-lived and safe for concurrent
// reads; mutation should occur only during startup.
//
// Usage (abbreviated):
//
//	var app config.AppConfig
//	app.InProduction = os.Getenv("ENV") == "production"
//	app.InfoLog = log.New(os.Stdout, "INFO:\t", log.Ldate|log.Ltime)
//	app.ErrorLog = log.New(os.Stdout, "ERROR:\t", log.Ldate|log.Ltime|log.Lshortfile)
//	app.Session = scs.New()
//	app.TemplateCache = make(map[string]*template.Template)
//	app.UseCache = app.InProduction
//	app.MailChan = make(chan models.MailData)
type AppConfig struct {
	// UseCache controls whether the renderer uses a prebuilt template cache.
	// In production this is typically true; in development it may be false to
	// force reparsing on each request.
	UseCache bool

	// TemplateCache stores compiled templates keyed by filename.
	// Populated at startup or lazily during development, depending on UseCache.
	TemplateCache map[string]*template.Template

	// InfoLog writes informational (non-error) operational messages.
	// Configure with a suitable io.Writer (e.g., stdout) and desired flags.
	InfoLog *log.Logger

	// ErrorLog writes error diagnostics. Prefer a destination that is collected
	// by your logging/observability pipeline. Include Lshortfile for context.
	ErrorLog *log.Logger

	// InProduction toggles production-only behaviors (e.g., Secure cookies, stricter
	// CSRF, disabled debug features). Set once at startup based on environment.
	InProduction bool

	// Session is the global session manager used across handlers. Configure cookie
	// attributes (lifetime, persistence, SameSite, Secure) during bootstrap.
	Session *scs.SessionManager

	// MailChan provides an asynchronous pathway for outbound mail work. A background
	// goroutine should drain this channel for the lifetime of the process.
	MailChan chan models.MailData
}
