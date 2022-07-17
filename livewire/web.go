package livewire

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"path"

	"github.com/gocopper/copper/cerrors"
	"github.com/gocopper/copper/chttp"
	"github.com/gocopper/copper/crandom"
)

var (
	//go:embed styles.html
	stylesHTML template.HTML

	//go:embed script.html
	scriptHTML template.HTML
)

var (
	HTMLRenderFuncStyles = chttp.HTMLRenderFunc{
		Name: "livewireStyles",
		Func: func(r *http.Request, _ *chttp.HTMLRenderer) interface{} {
			return func() template.HTML { return stylesHTML }
		},
	}
	HTMLRenderFuncScript = chttp.HTMLRenderFunc{
		Name: "livewireScript",
		Func: func(r *http.Request, _ *chttp.HTMLRenderer) interface{} {
			return func() template.HTML { return scriptHTML }
		},
	}
)

type HTMLRenderFuncLivewire = chttp.HTMLRenderFunc

func ProvideHTMLRenderFuncLivewire(components []Component) HTMLRenderFuncLivewire {
	componentByName := make(map[string]Component)
	for i := range components {
		componentByName[components[i].Name()] = components[i]
	}

	return chttp.HTMLRenderFunc{
		Name: "livewire",
		Func: func(r *http.Request, html *chttp.HTMLRenderer) interface{} {
			return func(name string, initialParams map[string]interface{}) (template.HTML, error) {
				var id = crandom.GenerateRandomString(20)

				c, ok := componentByName[name]
				if !ok {
					return "", cerrors.New(nil, "component does not exist", map[string]interface{}{
						"name": name,
					})
				}

				props := c.InitialProps()

				out, err := c.Render(props, func(cp string, data map[string]interface{}) (template.HTML, error) {
					return html.Render(r, path.Join("livewire", cp), mergeMaps(initialParams, data, props.data))
				})
				if err != nil {
					return "", cerrors.New(err, "failed to execute html template", map[string]interface{}{
						"props": props,
					})
				}

				initialData, err := json.Marshal(map[string]interface{}{
					"fingerprint": Fingerprint{
						ID:               id,
						Name:             name,
						Locale:           "en",
						Path:             r.URL.Path,
						Method:           r.Method,
						InvalidationHash: "aaa",
					},
					"serverMemo": ServerMemo{
						HTMLHash:  htmlHash(out),
						Props:     props,
						PropsMeta: nil,
						Children:  nil,
						Errors:    nil,
					},
					"effects": EffectsRequest{},
				})
				if err != nil {
					return "", cerrors.New(err, "failed to marshal initial data", nil)
				}

				html, err := updateHTML(out, map[string]string{
					"wire:id":           id,
					"wire:initial-data": string(initialData),
				}, fmt.Sprintf("<!-- Livewire Component wire-end:%s -->", id))
				if err != nil {
					return "", cerrors.New(err, "failed to render html", nil)
				}

				return html, nil
			}
		},
	}
}
