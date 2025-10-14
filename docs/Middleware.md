# Middleware

## Overview

Middleware in Proteus provides a powerful way to process HTTP requests and responses in a pipeline-like fashion. Middleware functions execute in sequence and can modify requests, generate responses, or perform side effects like logging, authentication, and data transformation. Each middleware can choose to continue to the next middleware or terminate the request early.

## Middleware Concepts

### Middleware Function Signature

```go
func middleware(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
    // Pre-processing logic here

    next() // Call next middleware or handler

    // Post-processing logic here (after response)
}
```

### Middleware Execution Flow

```go
// Request flows through middleware in order:
// Request -> Middleware1 -> Middleware2 -> Handler -> Middleware2 (post) -> Middleware1 (post) -> Response

server.Use(middleware1)  // Executes first
server.Use(middleware2)  // Executes second
server.Get("/", handler) // Final handler

// Execution order:
// 1. middleware1 (pre-processing)
// 2. middleware2 (pre-processing)
// 3. handler
// 4. middleware2 (post-processing)
// 5. middleware1 (post-processing)
```

## Built-in Middleware

### JSON Parser Middleware

```go
// Parse JSON request bodies
server.Post("/api/users", proteus.JsonParser, func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // req.Body contains parsed JSON as map[string]interface{}
    userData := req.Body.(map[string]interface{})

    name := userData["name"].(string)
    email := userData["email"].(string)

    res.Status(proteus.Status201)
    res.Send("User created: " + name)
})

// Apply JSON parser globally
server.Use(proteus.JsonParser)
```

### URL Encoded Parser Middleware

```go
// Parse URL-encoded form data
server.Post("/submit", proteus.UrlEncoded, func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // req.Body contains parsed form data as map[string]interface{}
    formData := req.Body.(map[string]interface{})

    username := formData["username"].(string)
    message := formData["message"].(string)

    res.Send("Form submitted by: " + username)
})

// Apply globally for all form submissions
server.Use(proteus.UrlEncoded)
```

## Custom Middleware

### Logging Middleware

```go
// Simple request logging
func loggingMiddleware(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
    start := time.Now()

    fmt.Printf("[%s] %s %s", start.Format("2006-01-02 15:04:05"), req.Method, req.ResourcePath)

    next()

    duration := time.Since(start)
    fmt.Printf(" - %v\n", duration)
}

server.Use(loggingMiddleware)

// Enhanced logging with user agent and response status
func enhancedLoggingMiddleware(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
    start := time.Now()
    userAgent, _ := req.Headers.Get("User-Agent")

    // Store start time for later use
    req.Locals["startTime"] = start

    next()

    duration := time.Since(start)

    // In a real implementation, you'd need to capture the response status
    // This is simplified for demonstration
    fmt.Printf("[%s] %s %s - %s - %v\n",
        start.Format("2006-01-02 15:04:05"),
        req.Method,
        req.ResourcePath,
        userAgent,
        duration)
}
```

### Authentication Middleware

```go
// JWT Bearer token authentication
func authMiddleware(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
    authHeader, exists := req.Headers.Get("Authorization")

    if !exists {
        res.Status(proteus.Status401)
        res.AddHeader("WWW-Authenticate", "Bearer")
        res.Send("Authorization header required")
        return
    }

    if !strings.HasPrefix(authHeader, "Bearer ") {
        res.Status(proteus.Status401)
        res.Send("Invalid authorization format")
        return
    }

    token := authHeader[7:] // Remove "Bearer " prefix

    // Validate JWT token (simplified)
    user, err := validateJWTToken(token)
    if err != nil {
        res.Status(proteus.Status401)
        res.Send("Invalid token: " + err.Error())
        return
    }

    // Store user in request locals
    req.Locals["user"] = user
    req.Locals["token"] = token

    next()
}

// API key authentication
func apiKeyMiddleware(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
    apiKey, exists := req.Headers.Get("X-API-Key")

    if !exists {
        res.Status(proteus.Status401)
        res.Send("API key required")
        return
    }

    // Validate API key
    apiUser, valid := validateAPIKey(apiKey)
    if !valid {
        res.Status(proteus.Status401)
        res.Send("Invalid API key")
        return
    }

    req.Locals["apiUser"] = apiUser
    req.Locals["apiKey"] = apiKey

    next()
}

// Apply to protected routes
server.Get("/protected", authMiddleware, func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    user := req.Locals["user"].(User)
    res.Send("Welcome, " + user.Name)
})

type User struct {
    ID   string
    Name string
    Role string
}

func validateJWTToken(token string) (*User, error) {
    // Simplified JWT validation
    if token == "valid-jwt-token" {
        return &User{ID: "123", Name: "John Doe", Role: "user"}, nil
    }
    return nil, fmt.Errorf("invalid token")
}

func validateAPIKey(key string) (*User, bool) {
    // Simplified API key validation
    if key == "valid-api-key" {
        return &User{ID: "api-user", Name: "API User", Role: "api"}, true
    }
    return nil, false
}
```

### CORS Middleware

```go
func corsMiddleware(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
    // Get origin from request
    origin, hasOrigin := req.Headers.Get("Origin")

    // Set CORS headers
    if hasOrigin {
        // Check allowed origins
        allowedOrigins := []string{
            "http://localhost:3000",
            "https://myapp.com",
            "https://www.myapp.com",
        }

        originAllowed := false
        for _, allowed := range allowedOrigins {
            if origin == allowed {
                originAllowed = true
                break
            }
        }

        if originAllowed {
            res.AddHeader("Access-Control-Allow-Origin", origin)
        }
    } else {
        res.AddHeader("Access-Control-Allow-Origin", "*")
    }

    // Set other CORS headers
    res.AddHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    res.AddHeader("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
    res.AddHeader("Access-Control-Allow-Credentials", "true")
    res.AddHeader("Access-Control-Max-Age", "3600")

    // Handle preflight OPTIONS request
    if req.Method == "OPTIONS" {
        res.Status(proteus.Status200)
        res.Send("")
        return
    }

    next()
}

server.Use(corsMiddleware)
```

### Rate Limiting Middleware

```go
// Simple in-memory rate limiting
var rateLimiter = make(map[string][]time.Time)
var rateLimiterMutex sync.Mutex

func rateLimitMiddleware(requestsPerMinute int) func(*proteus.HttpRequest, *proteus.HttpResponse, func()) {
    return func(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
        clientIP := getClientIP(req)
        now := time.Now()

        rateLimiterMutex.Lock()
        defer rateLimiterMutex.Unlock()

        // Get request history for this IP
        requests, exists := rateLimiter[clientIP]
        if !exists {
            requests = []time.Time{}
        }

        // Remove requests older than 1 minute
        cutoff := now.Add(-time.Minute)
        validRequests := []time.Time{}
        for _, reqTime := range requests {
            if reqTime.After(cutoff) {
                validRequests = append(validRequests, reqTime)
            }
        }

        // Check rate limit
        if len(validRequests) >= requestsPerMinute {
            res.Status(proteus.Status429)
            res.AddHeader("X-RateLimit-Limit", fmt.Sprintf("%d", requestsPerMinute))
            res.AddHeader("X-RateLimit-Remaining", "0")
            res.AddHeader("X-RateLimit-Reset", fmt.Sprintf("%d", now.Add(time.Minute).Unix()))
            res.AddHeader("Retry-After", "60")
            res.Send("Rate limit exceeded")
            return
        }

        // Add current request and update store
        validRequests = append(validRequests, now)
        rateLimiter[clientIP] = validRequests

        // Add rate limit headers
        remaining := requestsPerMinute - len(validRequests)
        res.AddHeader("X-RateLimit-Limit", fmt.Sprintf("%d", requestsPerMinute))
        res.AddHeader("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
        res.AddHeader("X-RateLimit-Reset", fmt.Sprintf("%d", now.Add(time.Minute).Unix()))

        next()
    }
}

// Apply rate limiting: 100 requests per minute
server.Use(rateLimitMiddleware(100))

func getClientIP(req *proteus.HttpRequest) string {
    // Try to get real IP from headers (simplified)
    if xForwardedFor, exists := req.Headers.Get("X-Forwarded-For"); exists {
        return strings.Split(xForwardedFor, ",")[0]
    }

    if xRealIP, exists := req.Headers.Get("X-Real-IP"); exists {
        return xRealIP
    }

    // Fallback to a default (in real implementation, extract from connection)
    return "127.0.0.1"
}
```

## Middleware Patterns

### Error Handling Middleware

```go
// Global error handler (should be applied early)
func errorHandlerMiddleware(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
    defer func() {
        if r := recover(); r != nil {
            // Log the error
            fmt.Printf("Panic recovered: %v\n", r)

            // Send error response
            res.Status(proteus.Status500)
            res.AddHeader("Content-Type", "application/json")
            res.Send(`{
                "error": "Internal Server Error",
                "message": "An unexpected error occurred",
                "timestamp": "` + time.Now().Format(time.RFC3339) + `"
            }`)
        }
    }()

    next()
}

server.Use(errorHandlerMiddleware)
```

### Request ID Middleware

```go
func requestIDMiddleware(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
    // Generate or extract request ID
    requestID := generateUUID()

    // Check if client provided request ID
    if clientID, exists := req.Headers.Get("X-Request-ID"); exists {
        requestID = clientID
    }

    // Store in request locals
    req.Locals["requestID"] = requestID

    // Add to response headers
    res.AddHeader("X-Request-ID", requestID)

    next()
}

func generateUUID() string {
    // Simplified UUID generation
    return fmt.Sprintf("req_%d_%d", time.Now().Unix(), rand.Intn(10000))
}

server.Use(requestIDMiddleware)
```

### Security Headers Middleware

```go
func securityHeadersMiddleware(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
    // Prevent clickjacking
    res.AddHeader("X-Frame-Options", "DENY")

    // XSS protection
    res.AddHeader("X-XSS-Protection", "1; mode=block")

    // Content type sniffing protection
    res.AddHeader("X-Content-Type-Options", "nosniff")

    // HSTS (for HTTPS)
    res.AddHeader("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

    // Content Security Policy
    res.AddHeader("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'")

    // Referrer Policy
    res.AddHeader("Referrer-Policy", "strict-origin-when-cross-origin")

    // Remove server information
    res.AddHeader("Server", "Proteus")

    next()
}

server.Use(securityHeadersMiddleware)
```

## Conditional Middleware

### Path-Specific Middleware

```go
// Apply middleware only to specific paths
func pathMiddleware(path string, middleware func(*proteus.HttpRequest, *proteus.HttpResponse, func())) func(*proteus.HttpRequest, *proteus.HttpResponse, func()) {
    return func(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
        if strings.HasPrefix(req.ResourcePath, path) {
            middleware(req, res, next)
        } else {
            next()
        }
    }
}

// Apply authentication only to /api paths
server.Use(pathMiddleware("/api", authMiddleware))

// Apply different middleware based on path patterns
func conditionalMiddleware(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
    switch {
    case strings.HasPrefix(req.ResourcePath, "/admin"):
        adminAuthMiddleware(req, res, next)
    case strings.HasPrefix(req.ResourcePath, "/api"):
        apiAuthMiddleware(req, res, next)
    case strings.HasPrefix(req.ResourcePath, "/public"):
        // No authentication needed
        next()
    default:
        basicAuthMiddleware(req, res, next)
    }
}
```

### Method-Specific Middleware

```go
func methodMiddleware(method string, middleware func(*proteus.HttpRequest, *proteus.HttpResponse, func())) func(*proteus.HttpRequest, *proteus.HttpResponse, func()) {
    return func(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
        if req.Method == method {
            middleware(req, res, next)
        } else {
            next()
        }
    }
}

// Apply JSON parsing only to POST, PUT, PATCH requests
server.Use(func(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
    if req.Method == "POST" || req.Method == "PUT" || req.Method == "PATCH" {
        proteus.JsonParser(req, res, next)
    } else {
        next()
    }
})
```

## Middleware Composition

### Middleware Chains

```go
// Compose multiple middleware into a chain
func createMiddlewareChain(middlewares ...func(*proteus.HttpRequest, *proteus.HttpResponse, func())) func(*proteus.HttpRequest, *proteus.HttpResponse, func()) {
    return func(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
        // Create a chain of next functions
        index := 0

        var chainNext func()
        chainNext = func() {
            if index < len(middlewares) {
                current := middlewares[index]
                index++
                current(req, res, chainNext)
            } else {
                next()
            }
        }

        chainNext()
    }
}

// Create an authentication chain
authChain := createMiddlewareChain(
    requestIDMiddleware,
    rateLimitMiddleware(60),
    authMiddleware,
)

server.Use("/protected", authChain)
```

### Middleware Groups

```go
// API middleware group
func apiMiddlewareGroup() []func(*proteus.HttpRequest, *proteus.HttpResponse, func()) {
    return []func(*proteus.HttpRequest, *proteus.HttpResponse, func()){
        corsMiddleware,
        rateLimitMiddleware(1000),
        proteus.JsonParser,
        authMiddleware,
    }
}

// Admin middleware group
func adminMiddlewareGroup() []func(*proteus.HttpRequest, *proteus.HttpResponse, func()) {
    return []func(*proteus.HttpRequest, *proteus.HttpResponse, func()){
        securityHeadersMiddleware,
        rateLimitMiddleware(100),
        adminAuthMiddleware,
    }
}

// Apply middleware groups
for _, middleware := range apiMiddlewareGroup() {
    server.Use("/api", middleware)
}

for _, middleware := range adminMiddlewareGroup() {
    server.Use("/admin", middleware)
}
```

## Best Practices

### Middleware Order
1. **Error handling** - First to catch all panics
2. **Logging** - Early to capture all requests
3. **Security headers** - Before any content processing
4. **CORS** - Before authentication for preflight requests
5. **Rate limiting** - Before expensive operations
6. **Authentication** - Before business logic
7. **Body parsing** - After authentication
8. **Business logic** - Route handlers
9. **404 handler** - Last middleware

### Performance Considerations
- Keep middleware lightweight and fast
- Avoid heavy computations in middleware
- Use early returns to stop processing when appropriate
- Cache authentication results when possible
- Use efficient data structures for rate limiting

### Security Guidelines
- Validate all inputs in middleware
- Sanitize user data before processing
- Use secure defaults for security headers
- Implement proper error handling without leaking information
- Log security events for monitoring

### Testing Middleware
- Write unit tests for each middleware function
- Test middleware in isolation
- Test middleware combinations
- Verify middleware order effects
- Test error conditions and edge cases
