# Proteus

## Overview

Proteus is a lightweight, Express.js-inspired web framework for Go that enables rapid development of HTTP servers and web APIs. The framework provides an intuitive, Node.js-like experience while leveraging Go's performance and type safety. Proteus supports HTTP/0.9, HTTP/1.0, and HTTP/1.1 protocols with comprehensive routing, middleware, and request/response handling capabilities.

## Key Features

- **Express-like API**: Familiar routing and middleware patterns for developers coming from Node.js
- **Multi-Protocol Support**: HTTP/0.9, HTTP/1.0, and HTTP/1.1 compatibility
- **Flexible Routing**: Path parameters, query parameters, and pattern matching
- **Middleware System**: Composable middleware for authentication, parsing, logging, and more
- **Static File Serving**: Built-in static asset serving with automatic MIME detection
- **Request Parsing**: JSON and URL-encoded form data parsing middleware
- **Comprehensive Logging**: Multiple logging formats (Common, Dev, Tiny, Short)
- **Type Safety**: Full Go type system integration with zero external dependencies

## Quick Start

### Installation

```bash
go mod init your-project
go get github.com/citadelofcode/proteus
```

### Basic Usage

```go
package main

import (
    "github.com/citadelofcode/proteus"
)

func main() {
    // Create a new server instance
    server := proteus.CreateServer()

    // Define a simple route
    server.Get("/", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
        res.Send("Hello, World!")
    })

    // Define a route with parameters
    server.Get("/users/:id", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
        userID, _ := req.Segments.Get("id")
        res.Send("User ID: " + userID[0])
    })

    // JSON API endpoint
    server.Post("/api/users", proteus.JsonParser, func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
        userData := req.Body
        res.Status(proteus.Status201)
        res.Send("User created successfully")
    })

    // Start the server
    server.Listen(8080, "localhost")
}
```

### Advanced Usage with Routing and Middleware

```go
package main

import (
    "fmt"
    "github.com/citadelofcode/proteus"
)

func main() {
    server := proteus.CreateServer()

    // Create API router
    apiRouter := proteus.CreateRouter()

    // Authentication middleware
    authMiddleware := func(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
        authHeader, exists := req.Headers.Get("Authorization")
        if !exists || authHeader == "" {
            res.Status(proteus.Status401)
            res.Send("Unauthorized")
            return
        }
        next() // Continue to next middleware/handler
    }

    // Apply middleware to API routes
    apiRouter.Use(authMiddleware)
    apiRouter.Use(proteus.JsonParser)

    // Define API routes
    apiRouter.Get("/users", getUsersHandler)
    apiRouter.Post("/users", createUserHandler)
    apiRouter.Get("/users/:id", getUserHandler)
    apiRouter.Put("/users/:id", updateUserHandler)
    apiRouter.Delete("/users/:id", deleteUserHandler)

    // Mount API router
    server.Use("/api/v1", apiRouter)

    // Serve static files
    server.Static("/static", "/var/www/static")

    // Global error handling
    server.Use(func(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
        defer func() {
            if r := recover(); r != nil {
                res.Status(proteus.Status500)
                res.Send("Internal Server Error")
            }
        }()
        next()
    })

    fmt.Println("Server starting on http://localhost:3000")
    server.Listen(3000, "0.0.0.0")
}

func getUsersHandler(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // Implementation here
    res.Send("List of users")
}

func createUserHandler(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    userData := req.Body
    // Process user creation
    res.Status(proteus.Status201)
    res.Send("User created")
}

func getUserHandler(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    userID, _ := req.Segments.Get("id")
    res.Send("User: " + userID[0])
}

func updateUserHandler(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    userID, _ := req.Segments.Get("id")
    userData := req.Body
    // Update logic
    res.Send("User " + userID[0] + " updated")
}

func deleteUserHandler(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    userID, _ := req.Segments.Get("id")
    // Delete logic
    res.Status(proteus.Status204)
    res.Send("")
}
```

## Architecture

The Proteus framework is organized into several key components:

1. **Core Server** (`HttpServer`): Main server instance handling HTTP connections and routing
2. **Request/Response** (`HttpRequest`, `HttpResponse`): HTTP message handling and processing
3. **Router System** (`Router`): Modular routing with pattern matching and middleware support
4. **Parameter Handling** (`Params`): URL parameters and query string processing
5. **Header Management** (`Headers`): HTTP header manipulation and access
6. **File Operations** (`File`): Static file serving and file system integration
7. **Middleware System**: Composable request/response processing pipeline
8. **Status Codes** (`StatusCode`): Type-safe HTTP status code constants

## API Reference

### Core Components

- [HttpServer](./HttpServer.md) - Main server instance for handling HTTP requests
- [HttpRequest](./HttpRequest.md) - HTTP request representation and parsing
- [HttpResponse](./HttpResponse.md) - HTTP response building and transmission

### Routing and Parameters

- [Router](./Router.md) - Modular routing system with middleware support
- [Params](./Params.md) - URL and query parameter handling
- [Headers](./Headers.md) - HTTP header management

### File and Static Content

- [File](./File.md) - File system operations and static content serving

### Middleware and Utilities

- [Middleware](./Middleware.md) - Built-in middleware functions and custom middleware
- [StatusCodes](./StatusCodes.md) - HTTP status code constants and utilities
- [Functions](./Functions.md) - Server and router creation functions

## Error Handling

Proteus provides comprehensive error handling through proper HTTP status codes and response patterns:

```go
server.Get("/api/users/:id", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    userID, exists := req.Segments.Get("id")
    if !exists || len(userID) == 0 {
        res.Status(proteus.Status400)
        res.Send("Missing user ID")
        return
    }

    // Simulate user lookup
    user, err := findUser(userID[0])
    if err != nil {
        res.Status(proteus.Status500)
        res.Send("Internal server error")
        return
    }

    if user == nil {
        res.Status(proteus.Status404)
        res.Send("User not found")
        return
    }

    res.Status(proteus.Status200)
    res.Send("User found: " + user.Name)
})
```

## Requirements

- **Go 1.21+**: Modern Go version with generics support
- **Standard Library**: Uses only Go standard library packages
- **Zero Dependencies**: No external dependencies required

## Best Practices

### Route Organization

```go
// Group related routes using routers
userRouter := proteus.CreateRouter()
userRouter.Get("/", listUsers)
userRouter.Post("/", createUser)
userRouter.Get("/:id", getUser)
userRouter.Put("/:id", updateUser)
userRouter.Delete("/:id", deleteUser)

server.Use("/users", userRouter)
```

### Middleware Usage

```go
// Apply middleware in order of execution
server.Use(loggingMiddleware)     // First: log requests
server.Use(authMiddleware)        // Second: authenticate
server.Use(proteus.JsonParser)    // Third: parse JSON
server.Use(rateLimitMiddleware)   // Fourth: rate limiting
```

### Error Handling

```go
// Use consistent error response format
func sendError(res *proteus.HttpResponse, status proteus.StatusCode, message string) {
    res.Status(status)
    res.AddHeader("Content-Type", "application/json")
    res.Send(`{"error": "` + message + `"}`)
}
```
