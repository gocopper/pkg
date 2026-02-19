package inertia

import (
	"encoding/json"
	"html/template"
)

type (
	Page struct {
		Component     string              `json:"component"`
		Props         map[string]any      `json:"props"`
		MergeProps    []string            `json:"mergeProps,omitempty"`
		DeferredProps map[string][]string `json:"deferredProps,omitempty"`
		URL           string              `json:"url"`
		Version       string              `json:"version"`
	}

	SSRRenderResponse struct {
		Head json.RawMessage `json:"head"`
		Body template.HTML   `json:"body"`
	}
)
