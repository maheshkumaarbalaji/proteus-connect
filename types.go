package proteus

import (
	"github.com/citadelofcode/proteus/internal"
)

// HttpServer represents a web server instance that can accept and process incoming HTTP requests.
// It provides methods for configuring middleware, routing, and starting the server to listen
// on a specific port and host. Create instances using CreateServer() function.
//
// Example usage:
//
//	server := proteus.CreateServer()
//	server.Listen(8080, "localhost")
type HttpServer = internal.HttpServer

// HttpRequest represents an HTTP request received by the web server.
// It contains all the information about the incoming request including headers, body,
// URL parameters, query parameters, and metadata. This type is passed to route handlers
// and middleware functions as the first parameter.
//
// Key fields accessible through methods:
//   - Headers: Access request headers
//   - Body: Parsed request body (requires appropriate middleware)
//   - Params: URL path parameters and query parameters
//   - Method: HTTP method (GET, POST, etc.)
//   - URL: Request URL and path information
type HttpRequest = internal.HttpRequest

// HttpResponse represents an HTTP response that will be sent back to the client.
// It provides methods for setting response headers, status codes, and body content.
// This type is passed to route handlers and middleware functions as the second parameter.
//
// Common methods:
//   - Status(code): Set the HTTP status code
//   - Send(data): Send response data
//   - JSON(data): Send JSON response
//   - Header(key, value): Set response header
type HttpResponse = internal.HttpResponse

// Headers represents a collection of HTTP headers stored as key-value pairs.
// It provides methods for getting, setting, and manipulating HTTP headers
// for both requests and responses. Header names are case-insensitive.
//
// Example usage:
//
//	headers.Get("Content-Type")
//	headers.Set("Authorization", "Bearer token")
type Headers = internal.Headers

// Params represents a collection of parameters extracted from HTTP requests.
// This includes both URL path parameters (e.g., /users/:id) and query string parameters
// (e.g., ?name=value&page=1). All parameter values are stored as strings.
//
// Example usage:
//
//	userID := params.Get("id")        // URL parameter
//	page := params.Get("page")        // Query parameter
type Params = internal.Params

// StatusCode represents an HTTP status code that can be sent in responses.
// Use the predefined status code constants (Status200, Status404, etc.) rather
// than raw integer values for better code readability and type safety.
//
// Example usage:
//
//	response.Status(proteus.Status200)
//	response.Status(proteus.Status404)
type StatusCode = internal.StatusCode

// Router provides functionality for defining HTTP routes and their associated handlers.
// It supports HTTP methods (GET, POST, PUT, DELETE, etc.), middleware application,
// and nested routing. Routers can be mounted on servers or other routers.
//
// Example usage:
//
//	router := proteus.CreateRouter()
//	router.Get("/users", getUsersHandler)
//	router.Post("/users", createUserHandler)
//	server.Use(router)
type Router = internal.Router

// File represents a single file from the local filesystem that can be accessed
// through the proteus framework. This is typically used for file upload handling
// and static file serving functionality.
//
// Common properties accessible:
//   - Name: Original filename
//   - Size: File size in bytes
//   - Content: File content data
type File = internal.File
