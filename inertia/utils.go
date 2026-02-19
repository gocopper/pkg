package inertia

import (
	"net/http"
	"slices"
	"strings"
)

// ShouldLoadProp returns true if any of the given props should be loaded for this request.
// Returns true for full renders and for partial reloads that include at least one of the props.
func ShouldLoadProp(req *http.Request, prop ...string) bool {
	partialData := req.Header.Get("x-inertia-partial-data")
	if partialData == "" {
		return true
	}
	return anyPropInPartialData(partialData, prop)
}

// ShouldLoadDeferredProp returns true if any of the given deferred props should be loaded for this request.
// Returns false for full renders (deferred props are skipped on initial load) and true only
// for partial reloads that include at least one of the props.
func ShouldLoadDeferredProp(req *http.Request, prop ...string) bool {
	partialData := req.Header.Get("x-inertia-partial-data")
	if partialData == "" {
		return false
	}
	return anyPropInPartialData(partialData, prop)
}

func anyPropInPartialData(partialData string, props []string) bool {
	for _, p := range strings.Split(partialData, ",") {
		if slices.Contains(props, p) {
			return true
		}
	}
	return false
}

// trimBasePathFromURL removes the base path prefix from a URL.
// Returns "/" if the result would be empty.
func trimBasePathFromURL(url, basePath string) string {
	trimmed := strings.TrimPrefix(url, basePath)
	if trimmed == "" {
		return "/"
	}
	return trimmed
}
