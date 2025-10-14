# Functions

## Overview

The Functions module in Proteus provides the core factory functions for creating server and router instances. These are the entry points for building web applications with the Proteus framework. The functions are designed to be simple, intuitive, and follow the Express.js-inspired API pattern.

## Core Functions

### CreateServer

The `CreateServer` function creates a new HTTP server instance that can handle incoming requests, define routes, and manage middleware.

```go
import "github.com/citadelofcode/proteus"

// Create a new server instance
server := proteus.CreateServer()

// The server is ready to use immediately
server.Get("/", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    res.Send("Hello, World!")
})

server.Listen(8080, "localhost")
```

#### Server Capabilities

```go
server := proteus.CreateServer()

// HTTP method handlers
server.Get("/users", getUsersHandler)
server.Post("/users", createUserHandler)
server.Put("/users/:id", updateUserHandler)
server.Delete("/users/:id", deleteUserHandler)
server.Patch("/users/:id", patchUserHandler)
server.Head("/users/:id", headUserHandler)
server.Options("/users", optionsHandler)

// Middleware support
server.Use(loggingMiddleware)
server.Use(authMiddleware)

// Static file serving
server.Static("/assets", "./public/assets")

// Router mounting
apiRouter := proteus.CreateRouter()
server.Use("/api/v1", apiRouter)

// Server configuration
server.SetLogger(proteus.INFO_LEVEL)
server.Listen(3000, "0.0.0.0")
```

### CreateRouter

The `CreateRouter` function creates a new router instance that can be used to organize related routes and middleware. Routers can be mounted on servers or nested within other routers.

```go
// Create a new router instance
router := proteus.CreateRouter()

// Define routes on the router
router.Get("/", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    res.Send("Router home")
})

router.Get("/:id", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    id, _ := req.Segments.Get("id")
    res.Send("Item ID: " + id[0])
})

// Mount router on server
server := proteus.CreateServer()
server.Use("/items", router)
```

#### Router Capabilities

```go
apiRouter := proteus.CreateRouter()

// All HTTP methods supported
apiRouter.Get("/status", statusHandler)
apiRouter.Post("/data", createDataHandler)
apiRouter.Put("/data/:id", updateDataHandler)
apiRouter.Delete("/data/:id", deleteDataHandler)

// Router-specific middleware
apiRouter.Use(apiAuthMiddleware)
apiRouter.Use(proteus.JsonParser)

// Static file serving on router
apiRouter.Static("/docs", "./api-documentation")

// Nested routers
userRouter := proteus.CreateRouter()
userRouter.Get("/profile", profileHandler)
apiRouter.Use("/users", userRouter)

// Mount on server
server.Use("/api", apiRouter)
```

## Function Usage Patterns

### Single Server Application

```go
package main

import "github.com/citadelofcode/proteus"

func main() {
    // Create single server for simple applications
    server := proteus.CreateServer()

    // Global middleware
    server.Use(loggingMiddleware)
    server.Use(corsMiddleware)

    // Routes
    server.Get("/", homeHandler)
    server.Get("/about", aboutHandler)
    server.Post("/contact", contactHandler)

    // Static files
    server.Static("/static", "./public")

    // Start server
    server.Listen(8080, "localhost")
}

func homeHandler(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    res.Send("Welcome to my website!")
}

func aboutHandler(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    res.Send("About us page")
}

func contactHandler(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    res.Send("Contact form submitted")
}

func loggingMiddleware(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
    fmt.Printf("%s %s\n", req.Method, req.ResourcePath)
    next()
}

func corsMiddleware(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
    res.AddHeader("Access-Control-Allow-Origin", "*")
    next()
}
```

### Modular Application with Routers

```go
package main

import (
    "github.com/citadelofcode/proteus"
)

func main() {
    server := proteus.CreateServer()

    // Global middleware
    server.Use(requestLoggingMiddleware)
    server.Use(errorHandlerMiddleware)

    // Create specialized routers
    authRouter := createAuthRouter()
    apiRouter := createAPIRouter()
    adminRouter := createAdminRouter()

    // Mount routers
    server.Use("/auth", authRouter)
    server.Use("/api", apiRouter)
    server.Use("/admin", adminRouter)

    // Static content
    server.Static("/", "./public")

    // Default route
    server.Get("/", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
        res.Send("Welcome to the application!")
    })

    server.Listen(8080, "0.0.0.0")
}

func createAuthRouter() *proteus.Router {
    router := proteus.CreateRouter()

    // JSON parsing for auth endpoints
    router.Use(proteus.JsonParser)

    router.Post("/login", loginHandler)
    router.Post("/register", registerHandler)
    router.Post("/logout", logoutHandler)
    router.Get("/verify", verifyTokenHandler)

    return router
}

func createAPIRouter() *proteus.Router {
    router := proteus.CreateRouter()

    // API middleware
    router.Use(authenticationMiddleware)
    router.Use(proteus.JsonParser)
    router.Use(rateLimitMiddleware)

    // Version 1 routes
    v1Router := proteus.CreateRouter()
    v1Router.Get("/users", v1GetUsersHandler)
    v1Router.Post("/users", v1CreateUserHandler)

    // Version 2 routes
    v2Router := proteus.CreateRouter()
    v2Router.Get("/users", v2GetUsersHandler)
    v2Router.Post("/users", v2CreateUserHandler)

    // Mount version routers
    router.Use("/v1", v1Router)
    router.Use("/v2", v2Router)

    return router
}

func createAdminRouter() *proteus.Router {
    router := proteus.CreateRouter()

    // Admin authentication
    router.Use(adminAuthMiddleware)
    router.Use(proteus.JsonParser)

    router.Get("/dashboard", adminDashboardHandler)
    router.Get("/users", adminUsersHandler)
    router.Delete("/users/:id", adminDeleteUserHandler)
    router.Get("/logs", adminLogsHandler)

    return router
}
```

## Function Reference

### CreateServer()
- **Returns**: `*proteus.HttpServer`
- **Description**: Creates a new HTTP server instance
- **Usage**: Entry point for creating web applications
- **Features**: Supports all HTTP methods, middleware, static files, router mounting

### CreateRouter()
- **Returns**: `*proteus.Router`
- **Description**: Creates a new router instance for organizing routes
- **Usage**: Modular route organization, nested routing, middleware grouping
- **Features**: Same route methods as server, can be mounted on servers or other routers

### Best Practices
1. **Use CreateServer()** for the main application entry point
2. **Use CreateRouter()** for organizing related routes and middleware
3. **Mount routers** to create clean, modular application structure
4. **Apply middleware** at appropriate levels (global, router, or route-specific)
5. **Test functions** to ensure proper instantiation and configuration
