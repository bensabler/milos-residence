// Package config defines application-wide configuration and shared dependencies.
package config

import (
	"html/template"
	"log"

	"github.com/alexedwards/scs/v2"
)

// AppConfig holds runtime configuration and shared services in a central struct.
// Pass a pointer to this struct where needed to avoid global state.
type AppConfig struct {
	// UseCache toggles on-disk/in-memory template caching (true in production).
	UseCache bool
	// TemplateCache maps template names to parsed template sets.
	TemplateCache map[string]*template.Template
	// InfoLog is an optional logger for informational messages.
	InfoLog  *log.Logger
	ErrorLog *log.Logger
	// InProduction enables production-only behaviors (e.g., secure cookies).
	InProduction bool
	// Session is the session manager used across requests.
	Session *scs.SessionManager
}
