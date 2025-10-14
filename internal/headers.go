package internal

import (
	"net/textproto"
	"strings"
)

// Represents a collection of headers (request or response).
type Headers map[string][]string

// Add inserts header key-value pair with automatic MIME key normalization and multi-value support.
// Use to add HTTP headers like Content-Type, Set-Cookie, or Accept with proper formatting.
// Normalizes keys to canonical format, splits comma-separated values, appends to existing headers.
// Parameters: key (header name), value (header value, comma-separated values split automatically).
// Returns: No return value, header stored in collection for HTTP transmission.
func (headers Headers) Add(key string, value string) {
	key = textproto.CanonicalMIMEHeaderKey(key)
	valueParts := strings.Split(value, ",")
	_, ok := headers[key]
	if ok {
		headers[key] = append(headers[key], valueParts...)
	} else {
		headers[key] = valueParts
	}
}

// Get retrieves header value for specified key with case-insensitive lookup and multi-value joining.
// Use to check for Content-Type, Authorization, or any header with automatic key normalization.
// Joins multiple values with commas, normalizes key to canonical MIME format automatically.
// Parameter: key (header name, case-insensitive).
// Returns: string (combined header value), bool (true if header exists, false otherwise).
func (headers Headers) Get(key string) (string, bool) {
	key = textproto.CanonicalMIMEHeaderKey(key)
	valueParts, ok := headers[key]
	if ok {
		return strings.Join(valueParts, ","), true
	} else {
		return "", false
	}
}

// Length returns total number of unique header keys in the headers collection.
// Use for validation, debugging, logging header collection size, and implementing header limits.
// Counts unique header keys only, not individual values (multi-value headers count as one).
// Parameters: No parameters.
// Returns: int (number of unique header keys in collection).
func (headers Headers) Length() int {
	return len(headers)
}
