package http

import (
	"strings"
)

// Collection of all parameters in a URL. The params can either be query parameters or path parameters.
type Params map[string][]string

// Returns the list of values in the map for the given key.
func (pr Params) Get(key string) ([]string, bool) {
	key = strings.TrimSpace(key)
	values, ok := pr[key]
	return values, ok
}

// Adds the given key-values pair to the params collection.
func (pr Params) Add(key string, paramValues []string) {
	key = strings.TrimSpace(key)
	values := make([]string, 0)
	values = append(values, paramValues...)
	pr[key] = values
}