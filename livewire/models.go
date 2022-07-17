package livewire

import (
	"encoding/json"
	"html/template"
)

type Message struct {
	Fingerprint Fingerprint `json:"fingerprint"`
	ServerMemo  ServerMemo  `json:"serverMemo"`
	Updates     []Update    `json:"updates"`
}

type MessageResponse struct {
	Effects    EffectsResponse `json:"effects"`
	ServerMemo ServerMemo      `json:"serverMemo"`
}

type Fingerprint struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Locale           string `json:"locale"`
	Path             string `json:"path"`
	Method           string `json:"method"`
	InvalidationHash string `json:"v"`
}

type EffectsRequest struct {
	Listeners []string `json:"listeners"`
}

type EffectsResponse struct {
	Dirty []string      `json:"dirty"`
	HTML  template.HTML `json:"html"`
}

type ServerMemo struct {
	HTMLHash  string   `json:"htmlHash"`
	Props     *Props   `json:"data"`
	PropsMeta []string `json:"dataMeta"`
	Children  []string `json:"children"`
	Errors    []string `json:"errors"`
}

type Update struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type UpdatePayloadCallMethod struct {
	ID     string   `json:"id"`
	Method string   `json:"method"`
	Params []string `json:"params"`
}

type UpdatePayloadSyncInput struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}
