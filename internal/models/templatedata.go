// Package models also provides view-model structs used to pass data from
// handlers to templates. TemplateData is the standard envelope inserted into
// all page renders by the render package.
package models

import "github.com/bensabler/milos-residence/internal/forms"

// TemplateData is the canonical container for data passed to HTML templates.
// It includes generic maps for ad hoc values, per-request UI messages, CSRF
// token, an optional form wrapper, and an auth marker used for conditional UI.
type TemplateData struct {
	StringMap       map[string]string      // Arbitrary string values by key
	IntMap          map[string]int         // Arbitrary int values by key
	FloatMap        map[string]float32     // Arbitrary float values by key
	Data            map[string]interface{} // Generic payload for complex views
	CSRFToken       string                 // CSRF token provided by middleware
	Flash           string                 // One-time success/info message
	Warning         string                 // One-time warning message
	Error           string                 // One-time error message
	Form            *forms.Form            // Optional form state/validation
	IsAuthenticated int                    // 1 if user is authenticated; else 0
}
