package inertia

import (
	"encoding/json"
	"html/template"
)

type (
	Page struct {
		Component  string         `json:"component"`
		Props      map[string]any `json:"props"`
		MergeProps []string       `json:"mergeProps"`
		URL        *string        `json:"url,omitempty"`
		Version    string         `json:"version"`
	}

	SSRRenderResponse struct {
		Head json.RawMessage `json:"head"`
		Body template.HTML   `json:"body"`
	}
)
