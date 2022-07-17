package livewire

import (
	"encoding/json"
	"html/template"
	"net/http"
	"path"
	"reflect"

	"github.com/gocopper/copper/cerrors"
	"github.com/gocopper/copper/chttp"
)

func NewRenderer(components []Component, html *chttp.HTMLRenderer) *Renderer {
	r := Renderer{
		componentByName: make(map[string]Component, len(components)),
		html:            html,
	}

	for i := range components {
		r.componentByName[components[i].Name()] = components[i]
	}

	return &r
}

type Renderer struct {
	componentByName map[string]Component
	html            *chttp.HTMLRenderer
}

func (r *Renderer) UpdateComponent(message *Message) (*MessageResponse, error) {
	c, ok := r.componentByName[message.Fingerprint.Name]
	if !ok {
		return nil, cerrors.New(nil, "component does not exist", map[string]interface{}{
			"name": message.Fingerprint.Name,
		})
	}

	componentVal := reflect.ValueOf(c)
	props := message.ServerMemo.Props

	for i := range message.Updates {
		update := message.Updates[i]

		switch update.Type {
		case "callMethod":
			var payload UpdatePayloadCallMethod
			err := json.Unmarshal(update.Payload, &payload)
			if err != nil {
				return nil, cerrors.New(err, "failed to unmarshal payload", map[string]interface{}{
					"type":    update.Type,
					"payload": string(update.Payload),
				})
			}

			ret := componentVal.MethodByName(payload.Method).Call([]reflect.Value{
				reflect.ValueOf(props),
			})

			if !ret[0].IsNil() {
				err = ret[0].Interface().(error)
				if err != nil {
					return nil, cerrors.New(err, "failed to call method on component", map[string]interface{}{
						"component": message.Fingerprint.Name,
						"payload":   payload,
					})
				}
			}
		case "syncInput":
			var payload UpdatePayloadSyncInput
			err := json.Unmarshal(update.Payload, &payload)
			if err != nil {
				return nil, cerrors.New(err, "failed to unmarshal payload", map[string]interface{}{
					"type":    update.Type,
					"payload": string(update.Payload),
				})
			}

			props.data[payload.Name] = payload.Value
		default:
			return nil, cerrors.New(nil, "unknown update type", map[string]interface{}{
				"type": update.Type,
			})
		}

	}

	err := c.OnUpdate(props)
	if err != nil {
		return nil, cerrors.New(err, "failed to call OnUpdate on component", map[string]interface{}{
			"component": message.Fingerprint.Name,
		})
	}

	initialReq, err := http.NewRequest(message.Fingerprint.Method, message.Fingerprint.Path, nil)
	if err != nil {
		return nil, cerrors.New(err, "failed to make initial http request", map[string]interface{}{
			"fingerprint": message.Fingerprint,
		})
	}

	out, err := c.Render(props, func(cp string, data map[string]interface{}) (template.HTML, error) {
		return r.html.Render(initialReq, path.Join("livewire", cp), mergeMaps(data, props.data))
	})
	if err != nil {
		return nil, cerrors.New(err, "failed to execute html template", map[string]interface{}{
			"props": props,
		})
	}
	updatedHTMLHash := htmlHash(out)

	effects := EffectsResponse{
		Dirty: make([]string, 0),
	}
	if message.ServerMemo.HTMLHash != updatedHTMLHash {
		html, err := updateHTML(out, map[string]string{
			"wire:id": message.Fingerprint.ID,
		}, "")
		if err != nil {
			return nil, cerrors.New(err, "failed to render html", nil)
		}

		effects.HTML = html
		effects.Dirty = props.dirty
	}

	return &MessageResponse{
		Effects: effects,
		ServerMemo: ServerMemo{
			HTMLHash: updatedHTMLHash,
			Props:    props,
		},
	}, nil
}
