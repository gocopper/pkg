package livewire

import (
	"net/http"

	"github.com/gocopper/copper/chttp"
	"github.com/gocopper/copper/clogger"
)

type NewRouterParams struct {
	RW       *chttp.ReaderWriter
	Renderer *Renderer
	Logger   clogger.Logger
}

func NewRouter(p NewRouterParams) *Router {
	return &Router{
		rw:       p.RW,
		renderer: p.Renderer,
		logger:   p.Logger,
	}
}

type Router struct {
	rw       *chttp.ReaderWriter
	renderer *Renderer
	logger   clogger.Logger
}

func (ro *Router) Routes() []chttp.Route {
	return []chttp.Route{
		{
			Path:    "/livewire/message/{component}",
			Methods: []string{http.MethodPost},
			Handler: ro.HandleLivewireMessage,
		},
	}
}

func (ro *Router) HandleLivewireMessage(w http.ResponseWriter, r *http.Request) {
	var message Message

	if ok := ro.rw.ReadJSON(w, r, &message); !ok {
		return
	}

	resp, err := ro.renderer.UpdateComponent(&message)
	if err != nil {
		ro.logger.Error("Failed to render livewire component update", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ro.rw.WriteJSON(w, chttp.WriteJSONParams{
		Data: &resp,
	})
}
