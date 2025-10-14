# HttpServer

## Overview

The `HttpServer` is the core component of the Proteus framework, responsible for handling incoming HTTP connections, routing requests, and managing the server lifecycle. It provides an Express.js-like API for defining routes, applying middleware, and serving content.

## Creation

```go
import "github.com/citadelofcode/proteus"

// Create a new server instance
server := proteus.CreateServer()
```

## HTTP Methods

The HttpServer supports all standard HTTP methods with dedicated handler functions:

### GET Requests

```go
// Simple GET route
server.Get("/", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    res.Send("Welcome to Proteus!")
})

// GET with path parameters
server.Get("/users/:id", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    userID, _ := req.Segments.Get("id")
    res.Send("User ID: " + userID[0])
})

// GET with query parameters
server.Get("/search", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    query, exists := req.Params.Get("q")
    if exists && len(query) > 0 {
        res.Send("Searching for: " + query[0])
    } else {
        res.Send("No search query provided")
    }
})
```

### POST Requests

```go
// POST with JSON middleware
server.Post("/api/users", proteus.JsonParser, func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    userData := req.Body.(map[string]interface{})
    name := userData["name"].(string)
    email := userData["email"].(string)

    // Create user logic here...

    res.Status(proteus.Status201)
    res.Send("User created: " + name)
})

// POST with form data
server.Post("/submit", proteus.UrlEncoded, func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    formData := req.Body.(map[string]interface{})
    username := formData["username"].(string)

    res.Send("Form submitted by: " + username)
})
```

### PUT, PATCH, DELETE

```go
// PUT for complete resource updates
server.Put("/api/users/:id", proteus.JsonParser, func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    userID, _ := req.Segments.Get("id")
    userData := req.Body.(map[string]interface{})

    // Update user completely
    res.Send("User " + userID[0] + " updated")
})

// PATCH for partial updates
server.Patch("/api/users/:id", proteus.JsonParser, func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    userID, _ := req.Segments.Get("id")
    updates := req.Body.(map[string]interface{})

    // Apply partial updates
    res.Send("User " + userID[0] + " patched")
})

// DELETE for resource removal
server.Delete("/api/users/:id", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    userID, _ := req.Segments.Get("id")

    // Delete user logic
    res.Status(proteus.Status204)
    res.Send("")
})
```

### Other HTTP Methods

```go
// HEAD for metadata only
server.Head("/api/users/:id", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    userID, _ := req.Segments.Get("id")
    // Set headers but no body
    res.AddHeader("Content-Type", "application/json")
    res.AddHeader("Content-Length", "156")
    res.Status(proteus.Status200)
})

// OPTIONS for CORS preflight
server.Options("/api/*", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    res.AddHeader("Access-Control-Allow-Origin", "*")
    res.AddHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    res.AddHeader("Access-Control-Allow-Headers", "Content-Type, Authorization")
    res.Status(proteus.Status200)
})
```

## Middleware

### Global Middleware

```go
// Apply middleware to all routes
server.Use(func(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
    // Logging middleware
    fmt.Printf("%s %s\n", req.Method, req.ResourcePath)
    next() // Continue to next middleware/handler
})

// Built-in middleware
server.Use(proteus.JsonParser)  // Parse JSON bodies globally
server.Use(proteus.UrlEncoded)  // Parse form data globally
```

### Route-Specific Middleware

```go
// Authentication middleware
authMiddleware := func(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
    token, exists := req.Headers.Get("Authorization")
    if !exists || token == "" {
        res.Status(proteus.Status401)
        res.Send("Unauthorized")
        return
    }
    // Validate token...
    next()
}

// Apply to specific route
server.Get("/protected", authMiddleware, func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    res.Send("This is protected content")
})

// Apply multiple middleware
server.Post("/api/admin", authMiddleware, adminOnlyMiddleware, proteus.JsonParser,
    func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
        res.Send("Admin endpoint")
    })
```

### Router Middleware

```go
// Mount routers with middleware
apiRouter := proteus.CreateRouter()
apiRouter.Use(proteus.JsonParser)
apiRouter.Use(authMiddleware)

server.Use("/api/v1", apiRouter)
```

## Static File Serving

```go
// Serve static files from a directory
server.Static("/static", "/var/www/static")
server.Static("/assets", "/path/to/assets")
server.Static("/uploads", "/var/uploads")

// Multiple static directories
server.Static("/css", "/assets/stylesheets")
server.Static("/js", "/assets/javascript")
server.Static("/images", "/assets/images")
```

## Server Lifecycle

### Starting the Server

```go
// Listen on specific port and host
server.Listen(8080, "localhost")    // Local development
server.Listen(3000, "0.0.0.0")      // All interfaces
server.Listen(80, "example.com")     // Production
```

### Logging Configuration

```go
// Set logging level
server.SetLogger(proteus.INFO_LEVEL)
server.SetLogger(proteus.WARN_LEVEL)
server.SetLogger(proteus.ERROR_LEVEL)

// The server automatically logs:
// - Request processing
// - Errors and warnings
// - Server startup/shutdown
```

## Advanced Configuration

### Custom Request Processing

```go
server.Use(func(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
    // Add custom properties to request
    req.Locals["timestamp"] = time.Now()
    req.Locals["requestId"] = generateUUID()

    // Custom headers
    res.AddHeader("X-Powered-By", "Proteus")
    res.AddHeader("X-Request-ID", req.Locals["requestId"].(string))

    next()
})
```

### Error Handling

```go
// Global error handler (should be last middleware)
server.Use(func(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
    defer func() {
        if r := recover(); r != nil {
            server.Log(proteus.ERROR_LEVEL, fmt.Sprintf("Panic recovered: %v", r))
            res.Status(proteus.Status500)
            res.Send("Internal Server Error")
        }
    }()
    next()
})
```

### CORS Handling

```go
// CORS middleware
corsMiddleware := func(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
    origin, _ := req.Headers.Get("Origin")

    res.AddHeader("Access-Control-Allow-Origin", origin)
    res.AddHeader("Access-Control-Allow-Credentials", "true")
    res.AddHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    res.AddHeader("Access-Control-Allow-Headers", "Content-Type, Authorization")

    if req.Method == "OPTIONS" {
        res.Status(proteus.Status200)
        res.Send("")
        return
    }

    next()
}

server.Use(corsMiddleware)
```

## Method Reference

### Route Definition Methods
- `Get(path, ...middleware, handler)` - Handle GET requests
- `Post(path, ...middleware, handler)` - Handle POST requests
- `Put(path, ...middleware, handler)` - Handle PUT requests
- `Patch(path, ...middleware, handler)` - Handle PATCH requests
- `Delete(path, ...middleware, handler)` - Handle DELETE requests
- `Head(path, ...middleware, handler)` - Handle HEAD requests
- `Options(path, ...middleware, handler)` - Handle OPTIONS requests

### Middleware and Routing
- `Use(...middleware)` - Apply middleware globally
- `Use(path, router)` - Mount router at path
- `Static(urlPath, fsPath)` - Serve static files

### Server Control
- `Listen(port, host)` - Start server listening
- `SetLogger(level)` - Set logging level
- `Log(level, message)` - Log message
