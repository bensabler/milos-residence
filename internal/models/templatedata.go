// Package models defines data structures shared between layers (e.g., templates).
package models

import "github.com/bensabler/milos-residence/internal/forms"

// TemplateData carries dynamic values from handlers to templates at render time.
type TemplateData struct {
	StringMap map[string]string      // small key/value strings for template use
	IntMap    map[string]int         // integer values (e.g., counters)
	FloatMap  map[string]float32     // float values (e.g., ratings)
	Data      map[string]interface{} // generic bag for ad-hoc payloads
	CSRFToken string                 // per-request token for POST form protection
	Flash     string                 // one-time success message
	Warning   string                 // one-time non-fatal warning
	Error     string                 // one-time error message
	Form      *forms.Form            // form state + validation messages
	// TODO(clean): prefer explicit typed fields over Data for long-lived values.
}
