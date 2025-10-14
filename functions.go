package proteus

import (
	"github.com/citadelofcode/proteus/internal"
)

// CreateServer creates a new HTTP server instance capable of accepting and processing HTTP requests.
// The server supports HTTP/0.9, HTTP/1.0, and HTTP/1.1 protocols and provides Express.js-like
// functionality including routing, middleware support, and request/response handling.
//
// Returns a pointer to an HttpServer instance that can be configured with routes and middleware
// before being started with the Listen() method.
//
// Example usage:
//
//	server := proteus.CreateServer()
//
//	// Configure routes
//	server.Get("/", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
//	    res.Send("Hello, World!")
//	})
//
//	// Start listening on port 8080
//	server.Listen(8080, "localhost")
//
// The server instance provides methods for:
//   - HTTP method handlers: Get(), Post(), Put(), Delete(), Patch(), etc.
//   - Middleware registration: Use()
//   - Static file serving: Static()
//   - Server lifecycle: Listen(), Close()
var CreateServer = internal.NewServer

// CreateRouter creates a new router instance for organizing and grouping HTTP route definitions.
// Routers provide a way to create modular, mountable route handlers that can be combined
// and reused across different parts of an application.
//
// Returns a pointer to a Router instance that can define routes and middleware independently
// from the main server instance. The router must be mounted on a server or another router
// using the Use() method to become functional.
//
// Example usage:
//
//	router := proteus.CreateRouter()
//
//	// Define routes on the router
//	router.Get("/users", getUsersHandler)
//	router.Post("/users", createUserHandler)
//	router.Get("/users/:id", getUserByIdHandler)
//
//	// Apply middleware to all routes in this router
//	router.Use(authenticationMiddleware)
//
//	// Mount the router on a server or another router
//	server.Use("/api", router)
//
// Router instances support:
//   - All HTTP method handlers: Get(), Post(), Put(), Delete(), etc.
//   - Middleware registration: Use()
//   - Route parameters: /users/:id, /posts/:slug
//   - Nested routing: mounting routers within other routers
var CreateRouter = internal.NewRouter
