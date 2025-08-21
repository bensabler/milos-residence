// Package models defines data structures shared between layers (e.g., templates).
package models

import "github.com/bensabler/milos-residence/internal/forms"

// TemplateData carries dynamic values from handlers to templates.
// Add or remove fields as your templates require.
type TemplateData struct {
	// StringMap holds string key/value pairs (e.g., labels or simple values).
	StringMap map[string]string
	// IntMap holds integer values.
	IntMap map[string]int
	// FloatMap holds float values.
	FloatMap map[string]float32
	// Data is a generic bag for arbitrary values. Prefer explicit fields when possible.
	Data map[string]interface{}
	// CSRFToken is a token for protecting POST forms against CSRF.
	CSRFToken string
	// Flash is for ephemeral success messages.
	Flash string
	// Warning is for non-fatal warnings.
	Warning string
	// Error is for fatal or error messages.
	Error string
	// Form is for checking a form after it's submitted
	Form *forms.Form
}
