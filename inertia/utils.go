package inertia

import "strings"

// trimBasePathFromURL removes the base path prefix from a URL.
// Returns "/" if the result would be empty.
func trimBasePathFromURL(url, basePath string) string {
	trimmed := strings.TrimPrefix(url, basePath)
	if trimmed == "" {
		return "/"
	}
	return trimmed
}
