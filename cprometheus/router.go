package cprometheus

import (
	"github.com/gocopper/copper/chttp"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

func NewRouter(config Config) *Router {
	return &Router{
		config:      config,
		promHandler: promhttp.Handler(),
	}
}

type Router struct {
	config      Config
	promHandler http.Handler
}

func (ro *Router) Routes() []chttp.Route {
	if !ro.config.HTTPEnabled {
		return []chttp.Route{}
	}

	path := "/internal/metrics"
	if ro.config.HTTPPath != "" {
		path = ro.config.HTTPPath
	}

	return []chttp.Route{
		{
			Path:    path,
			Methods: []string{http.MethodGet},
			Handler: ro.promHandler.ServeHTTP,
		},
	}
}
