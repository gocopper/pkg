package livewire

import (
	"html/template"
)

type Component interface {
	Name() string
	InitialProps() *Props
	OnUpdate(p *Props) error
	Render(p *Props, view ViewFn) (template.HTML, error)
}

type ViewFn = func(path string, data map[string]interface{}) (template.HTML, error)
