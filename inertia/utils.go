package inertia

import (
	"net/http"
	"strings"
)

// ShouldLoadProp returns true if the given prop should be loaded for this request.
// Returns true for full renders and for partial reloads that include the prop.
func ShouldLoadProp(req *http.Request, prop string) bool {
	partialData := req.Header.Get("x-inertia-partial-data")
	if partialData == "" {
		return true
	}
	for _, p := range strings.Split(partialData, ",") {
		if p == prop {
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
