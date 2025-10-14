package proteus

import (
	"github.com/citadelofcode/proteus/internal"
)

// JsonParser is middleware that automatically parses incoming JSON request payloads and makes
// the parsed data available in the request's Body field. This middleware should be applied
// to routes that expect to receive JSON data in the request body (typically POST, PUT, PATCH requests).
//
// The middleware:
//   - Checks for Content-Type: application/json header
//   - Parses the JSON payload from the request body
//   - Stores the parsed data in req.Body as a map[string]interface{}
//   - Handles parsing errors gracefully by returning appropriate HTTP error responses
//
// Usage:
//
//	// Apply to entire server
//	server.Use(proteus.JsonParser)
//
//	// Apply to specific router
//	router.Use(proteus.JsonParser)
//
//	// Apply to specific route
//	server.Post("/api/users", proteus.JsonParser, func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
//	    userData := req.Body // Contains parsed JSON data
//	    // Process the user data...
//	})
//
// Requirements:
//   - Request must have Content-Type: application/json header
//   - Request body must contain valid JSON
//   - Content-Length header should be present for best performance
var JsonParser = internal.JsonParser

// UrlEncoded is middleware that automatically parses URL-encoded form data from request payloads
// and makes the parsed data available in the request's Body field. This middleware is essential
// for handling HTML form submissions and other application/x-www-form-urlencoded content.
//
// The middleware:
//   - Checks for Content-Type: application/x-www-form-urlencoded header
//   - Parses form data from the request body
//   - Stores parsed data in req.Body as a map[string]interface{}
//   - Handles multiple values for the same field name
//   - Supports nested field names using bracket notation (field[subfield])
//
// Usage:
//
//	// Apply to entire server
//	server.Use(proteus.UrlEncoded)
//
//	// Apply to specific router
//	router.Use(proteus.UrlEncoded)
//
//	// Apply to specific route
//	server.Post("/submit-form", proteus.UrlEncoded, func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
//	    formData := req.Body // Contains parsed form data
//	    username := formData["username"]
//	    // Process form data...
//	})
//
// Requirements:
//   - Request must have Content-Type: application/x-www-form-urlencoded header
//   - Request body must contain valid URL-encoded data
//   - Content-Length header should be present for best performance
//
// Supported form field formats:
//   - Simple fields: name=value
//   - Multiple values: hobby=reading&hobby=gaming
//   - Nested fields: user[name]=john&user[age]=30
var UrlEncoded = internal.UrlEncoded
