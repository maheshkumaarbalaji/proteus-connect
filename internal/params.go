package internal

import (
	"strings"
)

// Collection of all parameters in a URL. The params can either be query parameters or path parameters.
type Params map[string][]string

// Get retrieves all values for specified parameter key from query and path parameters.
// Use to access route path parameters (:id), query parameters (?name=value), and multi-value parameters.
// Trims whitespace from key, preserves value order, handles both single and multiple values.
// Parameter: key (parameter name to lookup like "id", "name", "category").
// Returns: []string (slice of all values for parameter), bool (true if parameter exists).
func (pr Params) Get(key string) ([]string, bool) {
	key = strings.TrimSpace(key)
	values, ok := pr[key]
	if !ok {
		return nil, false
	}
	return values, true
}

// Add inserts or appends parameter values to params collection with multi-value support.
// Use for URL parsing, route matching, form processing, and middleware parameter injection.
// Creates new entry for new keys, appends to existing keys, preserves value order.
// Parameters: key (parameter name, whitespace trimmed), paramValues (slice of values to add).
// Returns: No return value, values stored in collection for retrieval via Get().
func (pr Params) Add(key string, paramValues []string) {
	key = strings.TrimSpace(key)
	values, ok := pr[key]
	if !ok {
		values = make([]string, 0)
	}
	values = append(values, paramValues...)
	pr[key] = values
}

// Length returns total number of unique parameter keys in the collection.
// Use for request validation, debugging, parameter limits, and monitoring request complexity.
// Counts unique parameter names only, not individual values (multi-value params count as one).
// Parameters: No parameters.
// Returns: int (number of unique parameter keys in collection).
func (pr Params) Length() int {
	return len(pr)
}
