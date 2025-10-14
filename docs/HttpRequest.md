# HttpRequest

## Overview

The `HttpRequest` object represents an incoming HTTP request to your Proteus server. It contains all the information about the client's request including headers, parameters, body data, and metadata. The request object is automatically passed as the first parameter to all route handlers.

## Basic Properties

### HTTP Method and Path

```go
server.Get("/api/users", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // Access the HTTP method
    method := req.Method // "GET"

    // Access the requested path
    path := req.ResourcePath // "/api/users"

    // Access the HTTP version
    version := req.Version // "HTTP/1.1"

    res.Send(fmt.Sprintf("Received %s request to %s via %s", method, path, version))
})
```

### Request Body

The request body is parsed based on the Content-Type header and available middleware:

```go
// JSON body parsing
server.Post("/api/users", proteus.JsonParser, func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // JSON data is parsed into map[string]interface{}
    userData := req.Body.(map[string]interface{})

    name := userData["name"].(string)
    email := userData["email"].(string)
    age := userData["age"].(float64)

    res.Send(fmt.Sprintf("Creating user: %s (%s), Age: %.0f", name, email, age))
})

// Form-encoded body parsing
server.Post("/submit", proteus.UrlEncoded, func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // Form data is parsed into map[string]interface{}
    formData := req.Body.(map[string]interface{})

    username := formData["username"].(string)
    message := formData["message"].(string)

    res.Send(fmt.Sprintf("Form submitted by %s: %s", username, message))
})

// Raw body access
server.Post("/webhook", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // Access raw body as string when no parser middleware is used
    rawBody := req.Body.(string)

    // Process raw data...
    res.Send("Webhook received")
})
```

## Headers

Access request headers using the Headers collection:

```go
server.Get("/info", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // Get specific headers
    userAgent, exists := req.Headers.Get("User-Agent")
    if exists {
        res.Send("Your browser: " + userAgent)
    }

    // Check for authorization header
    auth, hasAuth := req.Headers.Get("Authorization")
    if hasAuth && strings.HasPrefix(auth, "Bearer ") {
        token := auth[7:] // Remove "Bearer " prefix
        // Validate token...
    }

    // Get content type
    contentType, _ := req.Headers.Get("Content-Type")

    // Check if header exists
    if hasCustom := req.Headers.Has("X-Custom-Header"); hasCustom {
        custom, _ := req.Headers.Get("X-Custom-Header")
        res.Send("Custom header value: " + custom)
    }
})
```

### Common Header Patterns

```go
// Authentication
server.Use(func(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
    auth, exists := req.Headers.Get("Authorization")
    if !exists {
        res.Status(proteus.Status401)
        res.Send("Missing Authorization header")
        return
    }

    if !strings.HasPrefix(auth, "Bearer ") {
        res.Status(proteus.Status401)
        res.Send("Invalid Authorization format")
        return
    }

    token := auth[7:]
    // Validate token logic...
    req.Locals["user"] = getUserFromToken(token)
    next()
})

// Content negotiation
server.Get("/api/data", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    accept, _ := req.Headers.Get("Accept")

    data := map[string]interface{}{
        "id": 123,
        "name": "John Doe",
    }

    if strings.Contains(accept, "application/json") {
        res.AddHeader("Content-Type", "application/json")
        res.Send(fmt.Sprintf(`{"id": %d, "name": "%s"}`, data["id"], data["name"]))
    } else if strings.Contains(accept, "text/xml") {
        res.AddHeader("Content-Type", "text/xml")
        res.Send(fmt.Sprintf(`<user><id>%d</id><name>%s</name></user>`, data["id"], data["name"]))
    } else {
        res.Send(fmt.Sprintf("User: %s (ID: %d)", data["name"], data["id"]))
    }
})
```

## URL Parameters (Query String)

Access query string parameters using the Params collection:

```go
server.Get("/search", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // Single parameter
    query, exists := req.Params.Get("q")
    if exists && len(query) > 0 {
        searchTerm := query[0]
        res.Send("Searching for: " + searchTerm)
    } else {
        res.Send("No search query provided")
    }

    // Multiple values for same parameter
    tags, hasTags := req.Params.Get("tags")
    if hasTags {
        res.Send(fmt.Sprintf("Search tags: %v", tags))
    }

    // Check parameter existence
    if req.Params.Has("debug") {
        // Enable debug mode
    }

    // Pagination parameters
    page := "1"
    if pageParam, hasPage := req.Params.Get("page"); hasPage && len(pageParam) > 0 {
        page = pageParam[0]
    }

    limit := "10"
    if limitParam, hasLimit := req.Params.Get("limit"); hasLimit && len(limitParam) > 0 {
        limit = limitParam[0]
    }

    res.Send(fmt.Sprintf("Page: %s, Limit: %s", page, limit))
})

// Example URLs and their parameter access:
// /search?q=golang&tags=web&tags=framework&page=2&limit=20
// - req.Params.Get("q") returns (["golang"], true)
// - req.Params.Get("tags") returns (["web", "framework"], true)
// - req.Params.Get("page") returns (["2"], true)
// - req.Params.Get("missing") returns (nil, false)
```

## Path Segments (Route Parameters)

Access route parameters using the Segments collection:

```go
server.Get("/users/:id", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // Get path parameter
    userID, exists := req.Segments.Get("id")
    if exists && len(userID) > 0 {
        id := userID[0]
        res.Send("Fetching user with ID: " + id)
    } else {
        res.Status(proteus.Status400)
        res.Send("Invalid user ID")
    }
})

server.Get("/api/:version/users/:userID/posts/:postID", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // Multiple path parameters
    version, _ := req.Segments.Get("version")
    userID, _ := req.Segments.Get("userID")
    postID, _ := req.Segments.Get("postID")

    response := fmt.Sprintf("API v%s: User %s, Post %s",
        version[0], userID[0], postID[0])
    res.Send(response)
})

// Wildcard segments
server.Get("/files/*", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // Access wildcard content
    if wildcardPath, exists := req.Segments.Get("*"); exists && len(wildcardPath) > 0 {
        filePath := wildcardPath[0]
        res.Send("Requested file: " + filePath)
    }
})

// Example route matching:
// Route: /users/:id
// URL: /users/123
// req.Segments.Get("id") returns (["123"], true)
//
// Route: /api/:version/users/:userID/posts/:postID
// URL: /api/v2/users/456/posts/789
// req.Segments.Get("version") returns (["v2"], true)
// req.Segments.Get("userID") returns (["456"], true)
// req.Segments.Get("postID") returns (["789"], true)
```

## Local Variables

Store request-specific data using the Locals map:

```go
// Authentication middleware sets user data
authMiddleware := func(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
    token, _ := req.Headers.Get("Authorization")
    user := authenticateToken(token[7:]) // Remove "Bearer "

    // Store user in locals
    req.Locals["user"] = user
    req.Locals["authenticated"] = true
    req.Locals["requestTime"] = time.Now()

    next()
}

// Route handler accesses user data
server.Get("/profile", authMiddleware, func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // Access stored user data
    user := req.Locals["user"].(User)
    isAuth := req.Locals["authenticated"].(bool)
    requestTime := req.Locals["requestTime"].(time.Time)

    if isAuth {
        res.Send(fmt.Sprintf("Welcome back, %s! Request processed at %s",
            user.Name, requestTime.Format(time.RFC3339)))
    } else {
        res.Status(proteus.Status401)
        res.Send("Unauthorized")
    }
})

// Logging middleware uses locals for request tracking
server.Use(func(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
    // Generate request ID
    requestID := generateUUID()
    req.Locals["requestID"] = requestID
    req.Locals["startTime"] = time.Now()

    // Add to response headers
    res.AddHeader("X-Request-ID", requestID)

    next()

    // Log after request processing
    duration := time.Since(req.Locals["startTime"].(time.Time))
    fmt.Printf("Request %s completed in %v\n", requestID, duration)
})
```

## Content Type Detection

```go
server.Post("/upload", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    contentType, _ := req.Headers.Get("Content-Type")

    switch {
    case strings.Contains(contentType, "application/json"):
        // Handle JSON data
        jsonData := req.Body.(map[string]interface{})
        res.Send("Received JSON data")

    case strings.Contains(contentType, "application/x-www-form-urlencoded"):
        // Handle form data
        formData := req.Body.(map[string]interface{})
        res.Send("Received form data")

    case strings.Contains(contentType, "multipart/form-data"):
        // Handle file upload
        rawBody := req.Body.(string)
        res.Send("Received multipart data")

    case strings.Contains(contentType, "text/plain"):
        // Handle plain text
        textData := req.Body.(string)
        res.Send("Received text: " + textData)

    default:
        res.Status(proteus.Status415)
        res.Send("Unsupported Media Type")
    }
})
```

## Request Validation

```go
// Validation middleware
validateUserInput := func(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
    if req.Method == "POST" || req.Method == "PUT" {
        userData := req.Body.(map[string]interface{})

        // Check required fields
        if name, exists := userData["name"]; !exists || name == "" {
            res.Status(proteus.Status400)
            res.Send("Name is required")
            return
        }

        if email, exists := userData["email"]; !exists || !isValidEmail(email.(string)) {
            res.Status(proteus.Status400)
            res.Send("Valid email is required")
            return
        }

        // Store validated data
        req.Locals["validatedData"] = userData
    }

    next()
}

server.Post("/api/users", proteus.JsonParser, validateUserInput,
    func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
        // Use pre-validated data
        userData := req.Locals["validatedData"].(map[string]interface{})

        // Process user creation...
        res.Status(proteus.Status201)
        res.Send("User created successfully")
    })
```

## Property Reference

### Core Properties
- `Method` (string) - HTTP method (GET, POST, PUT, etc.)
- `ResourcePath` (string) - Requested URL path
- `Version` (string) - HTTP version (HTTP/1.0, HTTP/1.1, etc.)
- `Body` (interface{}) - Parsed request body

### Collections
- `Headers` (Headers) - HTTP request headers
- `Params` (Params) - URL query parameters
- `Segments` (Params) - Route path parameters
- `Locals` (map[string]interface{}) - Request-local storage

### Collection Methods
- `Get(key)` ([]string, bool) - Get values for key
- `Has(key)` bool - Check if key exists
- `Set(key, value)` - Set key-value pair (mainly for testing)
- `Delete(key)` - Remove key (mainly for testing)
