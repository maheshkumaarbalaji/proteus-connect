# Router

## Overview

The `Router` in Proteus provides a way to organize and group related routes together. It allows you to create modular route definitions that can be mounted on the main server or nested within other routers. This is essential for building scalable web applications with clean, organized routing structure.

## Creating Routers

```go
import "github.com/citadelofcode/proteus"

// Create a new router instance
router := proteus.CreateRouter()

// Create multiple routers for different purposes
apiRouter := proteus.CreateRouter()
adminRouter := proteus.CreateRouter()
userRouter := proteus.CreateRouter()
```

## Basic Router Usage

### Defining Routes on Router

```go
// Create API router
apiRouter := proteus.CreateRouter()

// Add routes to the router
apiRouter.Get("/status", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    res.AddHeader("Content-Type", "application/json")
    res.Send(`{"status": "API is running", "version": "1.0"}`)
})

apiRouter.Get("/health", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    res.AddHeader("Content-Type", "application/json")
    res.Send(`{"health": "OK", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`)
})

apiRouter.Post("/data", proteus.JsonParser, func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    data := req.Body.(map[string]interface{})
    res.AddHeader("Content-Type", "application/json")
    res.Send(fmt.Sprintf(`{"received": %+v}`, data))
})

// Mount router on server
server := proteus.CreateServer()
server.Use("/api/v1", apiRouter)
```

### Router HTTP Methods

Routers support all the same HTTP methods as the main server:

```go
userRouter := proteus.CreateRouter()

// GET routes
userRouter.Get("/", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    res.Send("User list")
})

userRouter.Get("/:id", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    userID, _ := req.Segments.Get("id")
    res.Send("User profile: " + userID[0])
})

// POST routes
userRouter.Post("/", proteus.JsonParser, func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    userData := req.Body.(map[string]interface{})
    res.Status(proteus.Status201)
    res.Send("User created: " + userData["name"].(string))
})

// PUT routes
userRouter.Put("/:id", proteus.JsonParser, func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    userID, _ := req.Segments.Get("id")
    userData := req.Body.(map[string]interface{})
    res.Send("User " + userID[0] + " updated")
})

// PATCH routes
userRouter.Patch("/:id", proteus.JsonParser, func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    userID, _ := req.Segments.Get("id")
    res.Send("User " + userID[0] + " partially updated")
})

// DELETE routes
userRouter.Delete("/:id", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    userID, _ := req.Segments.Get("id")
    res.Status(proteus.Status204)
    res.Send("")
})

// Mount on server at /users
server.Use("/users", userRouter)
```

## Router Middleware

### Router-Specific Middleware

```go
// API router with authentication middleware
apiRouter := proteus.CreateRouter()

// Apply middleware to all routes in this router
apiRouter.Use(func(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
    // API authentication
    apiKey, exists := req.Headers.Get("X-API-Key")
    if !exists || apiKey != "valid-api-key" {
        res.Status(proteus.Status401)
        res.AddHeader("Content-Type", "application/json")
        res.Send(`{"error": "Invalid API key"}`)
        return
    }

    req.Locals["authenticated"] = true
    next()
})

// Apply JSON parser to all routes
apiRouter.Use(proteus.JsonParser)

// Add routes that inherit the middleware
apiRouter.Post("/users", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // This route automatically has authentication and JSON parsing
    userData := req.Body.(map[string]interface{})
    res.Status(proteus.Status201)
    res.Send("User created with authentication")
})

server.Use("/api", apiRouter)
```

## Router Patterns

### RESTful Resource Routing

```go
// Create a RESTful router for a resource
func CreateResourceRouter(resourceName string) *proteus.Router {
    router := proteus.CreateRouter()

    // GET /resource - List all
    router.Get("/", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
        // Get query parameters for pagination
        page, hasPage := req.Params.Get("page")
        pageNum := "1"
        if hasPage && len(page) > 0 {
            pageNum = page[0]
        }

        res.AddHeader("Content-Type", "application/json")
        res.Send(fmt.Sprintf(`{"resource": "%s", "page": "%s", "items": []}`, resourceName, pageNum))
    })

    // GET /resource/:id - Get specific
    router.Get("/:id", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
        id, _ := req.Segments.Get("id")
        res.AddHeader("Content-Type", "application/json")
        res.Send(fmt.Sprintf(`{"resource": "%s", "id": "%s"}`, resourceName, id[0]))
    })

    // POST /resource - Create new
    router.Post("/", proteus.JsonParser, func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
        data := req.Body.(map[string]interface{})
        res.Status(proteus.Status201)
        res.AddHeader("Content-Type", "application/json")
        res.Send(fmt.Sprintf(`{"resource": "%s", "created": %+v}`, resourceName, data))
    })

    // PUT /resource/:id - Update complete
    router.Put("/:id", proteus.JsonParser, func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
        id, _ := req.Segments.Get("id")
        data := req.Body.(map[string]interface{})
        res.AddHeader("Content-Type", "application/json")
        res.Send(fmt.Sprintf(`{"resource": "%s", "id": "%s", "updated": %+v}`, resourceName, id[0], data))
    })

    // PATCH /resource/:id - Partial update
    router.Patch("/:id", proteus.JsonParser, func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
        id, _ := req.Segments.Get("id")
        data := req.Body.(map[string]interface{})
        res.AddHeader("Content-Type", "application/json")
        res.Send(fmt.Sprintf(`{"resource": "%s", "id": "%s", "patched": %+v}`, resourceName, id[0], data))
    })

    // DELETE /resource/:id - Remove
    router.Delete("/:id", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
        id, _ := req.Segments.Get("id")
        res.Status(proteus.Status204)
        res.Send("")
    })

    return router
}

// Usage
server.Use("/posts", CreateResourceRouter("posts"))
server.Use("/comments", CreateResourceRouter("comments"))
server.Use("/categories", CreateResourceRouter("categories"))
```

## Static File Routing

```go
// Static file router
staticRouter := proteus.CreateRouter()

// Serve different static content types
staticRouter.Static("/css", "/var/www/assets/stylesheets")
staticRouter.Static("/js", "/var/www/assets/javascript")
staticRouter.Static("/images", "/var/www/assets/images")
staticRouter.Static("/fonts", "/var/www/assets/fonts")

// Add cache headers for static content
staticRouter.Use(func(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
    // Long cache for static assets
    res.AddHeader("Cache-Control", "public, max-age=31536000") // 1 year
    res.AddHeader("Expires", time.Now().AddDate(1, 0, 0).Format(http.TimeFormat))

    next()
})

// Mount static router
server.Use("/assets", staticRouter)
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

### Middleware and Mounting
- `Use(...middleware)` - Apply middleware to router
- `Use(path, router)` - Mount sub-router at path
- `Static(urlPath, fsPath)` - Serve static files

### Best Practices
- Use routers to organize related functionality
- Apply authentication/authorization at router level
- Nest routers for complex applications
- Use consistent REST patterns
- Implement proper error handling per router
- Version APIs using separate routers
