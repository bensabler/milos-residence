package config

import (
	"html/template"
	"log"

	"github.com/alexedwards/scs/v2"
)

// AppConfig centralizes application-wide services and flags for dependency injection.
type AppConfig struct {
	UseCache      bool                          // toggle template caching (true in prod)
	TemplateCache map[string]*template.Template // parsed templates keyed by page name
	InfoLog       *log.Logger                   // operational/diagnostic logging
	ErrorLog      *log.Logger                   // error logging with stack traces
	InProduction  bool                          // production mode flag (affects cookies, security)
	Session       *scs.SessionManager           // HTTP session manager
	// TODO: consider adding AppName, BaseURL, and Version for templates and logs.
}
