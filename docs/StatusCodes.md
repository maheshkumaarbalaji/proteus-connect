# StatusCodes

## Overview

The `StatusCode` type in Proteus provides type-safe HTTP status codes that can be used with the `res.Status()` method. These constants ensure you're using valid HTTP status codes and improve code readability by providing descriptive names instead of raw numbers.

## Status Code Categories

HTTP status codes are organized into five categories based on their first digit:

- **1xx Informational** - Request received, continuing process
- **2xx Success** - Request successfully received, understood, and accepted
- **3xx Redirection** - Further action must be taken to complete the request
- **4xx Client Error** - Request contains bad syntax or cannot be fulfilled
- **5xx Server Error** - Server failed to fulfill an apparently valid request

## 1xx Informational Responses

```go
// Informational responses (rarely used in typical applications)
server.Get("/processing", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // 100 Continue - Client should continue with request
    res.Status(proteus.Status100)
    res.Send("Continue with your request")

    // 101 Switching Protocols - Server is switching protocols
    res.Status(proteus.Status101)
    res.Send("Switching protocols")

    // 102 Processing - Server has received and is processing the request
    res.Status(proteus.Status102)
    res.Send("Processing request")
})
```

## 2xx Success Responses

```go
// Successful responses
server.Get("/success-examples", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // 200 OK - Standard successful response
    res.Status(proteus.Status200)
    res.Send("Request successful")
})

server.Post("/create", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // 201 Created - Resource successfully created
    res.Status(proteus.Status201)
    res.Send("Resource created successfully")
})

server.Put("/update", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // 202 Accepted - Request accepted but not yet processed
    res.Status(proteus.Status202)
    res.Send("Request accepted for processing")
})

server.Delete("/delete/:id", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // 204 No Content - Success with no response body
    res.Status(proteus.Status204)
    res.Send("") // Empty body for 204
})

server.Get("/partial", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // 206 Partial Content - Partial resource delivered
    res.Status(proteus.Status206)
    res.AddHeader("Content-Range", "bytes 200-1023/2048")
    res.Send("Partial content here")
})
```

### Success Response Examples

```go
// API endpoint with different success scenarios
server.Post("/api/users", proteus.JsonParser, func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    userData := req.Body.(map[string]interface{})

    // Validate required fields
    name, hasName := userData["name"]
    email, hasEmail := userData["email"]

    if !hasName || !hasEmail {
        res.Status(proteus.Status400) // Bad Request
        res.Send("Name and email are required")
        return
    }

    // Create user (simulated)
    userID := createUser(name.(string), email.(string))

    // 201 Created - Resource successfully created
    res.Status(proteus.Status201)
    res.AddHeader("Content-Type", "application/json")
    res.AddHeader("Location", "/api/users/"+userID)
    res.Send(fmt.Sprintf(`{
        "id": "%s",
        "name": "%s",
        "email": "%s",
        "created": "%s"
    }`, userID, name, email, time.Now().Format(time.RFC3339)))
})

server.Put("/api/users/:id", proteus.JsonParser, func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    userID, _ := req.Segments.Get("id")
    userData := req.Body.(map[string]interface{})

    // Update user (simulated)
    updated := updateUser(userID[0], userData)

    if updated {
        // 200 OK - Resource successfully updated
        res.Status(proteus.Status200)
        res.Send("User updated successfully")
    } else {
        // 202 Accepted - Update queued for processing
        res.Status(proteus.Status202)
        res.Send("Update request accepted and queued")
    }
})

func createUser(name, email string) string {
    return fmt.Sprintf("user_%d", time.Now().Unix())
}

func updateUser(id string, data map[string]interface{}) bool {
    // Simulate immediate vs queued update
    return time.Now().Second()%2 == 0
}
```

## 3xx Redirection Responses

```go
// Redirection responses
server.Get("/redirect-examples", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // 301 Moved Permanently - Resource permanently moved
    res.Status(proteus.Status301)
    res.AddHeader("Location", "/new-location")
    res.Send("Resource moved permanently to /new-location")
})

server.Get("/old-page", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // 301 Moved Permanently
    res.Status(proteus.Status301)
    res.AddHeader("Location", "/new-page")
    res.Send("Page has moved permanently")
})

server.Get("/temporary-redirect", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // 302 Found (Temporary Redirect)
    res.Status(proteus.Status302)
    res.AddHeader("Location", "/temp-location")
    res.Send("Temporarily redirected")
})

server.Get("/see-other", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // 303 See Other - Redirect to different resource
    res.Status(proteus.Status303)
    res.AddHeader("Location", "/other-resource")
    res.Send("See other resource")
})

server.Get("/not-modified", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // Check if resource was modified
    ifNoneMatch, _ := req.Headers.Get("If-None-Match")
    currentETag := "\"abc123\""

    if ifNoneMatch == currentETag {
        // 304 Not Modified - Resource not changed
        res.Status(proteus.Status304)
        res.AddHeader("ETag", currentETag)
        res.Send("") // Empty body for 304
    } else {
        // 200 OK - Return full resource
        res.Status(proteus.Status200)
        res.AddHeader("ETag", currentETag)
        res.Send("Full resource content")
    }
})

server.Get("/temp-redirect-post", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // 307 Temporary Redirect - Method preserved
    res.Status(proteus.Status307)
    res.AddHeader("Location", "/new-endpoint")
    res.Send("Temporary redirect preserving method")
})

server.Get("/permanent-redirect-post", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // 308 Permanent Redirect - Method preserved
    res.Status(proteus.Status308)
    res.AddHeader("Location", "/new-permanent-endpoint")
    res.Send("Permanent redirect preserving method")
})
```

## 4xx Client Error Responses

```go
// Client error responses
server.Get("/client-errors", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // 400 Bad Request - Invalid request syntax
    res.Status(proteus.Status400)
    res.Send("Bad request - invalid syntax")
})

server.Get("/protected", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    auth, exists := req.Headers.Get("Authorization")

    if !exists {
        // 401 Unauthorized - Authentication required
        res.Status(proteus.Status401)
        res.AddHeader("WWW-Authenticate", "Bearer")
        res.Send("Authentication required")
        return
    }

    if !isValidToken(auth) {
        // 401 Unauthorized - Invalid credentials
        res.Status(proteus.Status401)
        res.Send("Invalid authentication credentials")
        return
    }

    res.Send("Access granted")
})

server.Get("/admin-only", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    user := getCurrentUser(req)

    if user == nil {
        // 401 Unauthorized
        res.Status(proteus.Status401)
        res.Send("Authentication required")
        return
    }

    if !user.IsAdmin {
        // 403 Forbidden - Access denied
        res.Status(proteus.Status403)
        res.Send("Access denied - admin rights required")
        return
    }

    res.Send("Admin content")
})

server.Get("/api/users/:id", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    userID, _ := req.Segments.Get("id")

    user := findUser(userID[0])
    if user == nil {
        // 404 Not Found - Resource doesn't exist
        res.Status(proteus.Status404)
        res.AddHeader("Content-Type", "application/json")
        res.Send(`{
            "error": "User not found",
            "message": "No user exists with the provided ID"
        }`)
        return
    }

    res.Send("User found: " + user.Name)
})

server.Post("/api/upload", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    contentType, _ := req.Headers.Get("Content-Type")

    if !strings.Contains(contentType, "multipart/form-data") {
        // 415 Unsupported Media Type
        res.Status(proteus.Status415)
        res.Send("Only multipart/form-data is supported for uploads")
        return
    }

    res.Send("File upload processed")
})

server.Post("/api/rate-limited", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    if isRateLimited(req) {
        // 429 Too Many Requests
        res.Status(proteus.Status429)
        res.AddHeader("Retry-After", "60")
        res.AddHeader("X-RateLimit-Limit", "100")
        res.AddHeader("X-RateLimit-Remaining", "0")
        res.Send("Rate limit exceeded. Please try again later.")
        return
    }

    res.Send("Request processed")
})

func isValidToken(token string) bool {
    return strings.HasPrefix(token, "Bearer valid-token")
}

func getCurrentUser(req *proteus.HttpRequest) *User {
    // Simplified user detection
    if auth, exists := req.Headers.Get("Authorization"); exists {
        if strings.Contains(auth, "admin-token") {
            return &User{Name: "Admin", IsAdmin: true}
        }
        return &User{Name: "User", IsAdmin: false}
    }
    return nil
}

func findUser(id string) *User {
    if id == "123" {
        return &User{Name: "John Doe", IsAdmin: false}
    }
    return nil
}

func isRateLimited(req *proteus.HttpRequest) bool {
    // Simplified rate limiting check
    return time.Now().Second()%10 == 0 // 10% chance of rate limit
}

type User struct {
    Name    string
    IsAdmin bool
}
```

### Common Client Error Patterns

```go
// Validation error with detailed response
server.Post("/api/validate", proteus.JsonParser, func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    data := req.Body.(map[string]interface{})
    errors := []string{}

    // Validate required fields
    if name, exists := data["name"]; !exists || name == "" {
        errors = append(errors, "Name is required")
    }

    if email, exists := data["email"]; !exists || !isValidEmail(email.(string)) {
        errors = append(errors, "Valid email is required")
    }

    if age, exists := data["age"]; exists {
        if ageFloat, ok := age.(float64); !ok || ageFloat < 0 || ageFloat > 150 {
            errors = append(errors, "Age must be between 0 and 150")
        }
    }

    if len(errors) > 0 {
        // 400 Bad Request with validation errors
        res.Status(proteus.Status400)
        res.AddHeader("Content-Type", "application/json")
        res.Send(fmt.Sprintf(`{
            "error": "Validation failed",
            "details": %s
        }`, formatErrors(errors)))
        return
    }

    res.Status(proteus.Status201)
    res.Send("Data validated and created")
})

func isValidEmail(email string) bool {
    return strings.Contains(email, "@") && strings.Contains(email, ".")
}

func formatErrors(errors []string) string {
    quotedErrors := make([]string, len(errors))
    for i, err := range errors {
        quotedErrors[i] = "\"" + err + "\""
    }
    return "[" + strings.Join(quotedErrors, ", ") + "]"
}
```

## 5xx Server Error Responses

```go
// Server error responses
server.Get("/server-error", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // 500 Internal Server Error - Generic server error
    res.Status(proteus.Status500)
    res.AddHeader("Content-Type", "application/json")
    res.Send(`{
        "error": "Internal Server Error",
        "message": "An unexpected error occurred"
    }`)
})

server.Get("/not-implemented", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // 501 Not Implemented - Feature not implemented
    res.Status(proteus.Status501)
    res.Send("This feature is not yet implemented")
})

server.Get("/bad-gateway", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // 502 Bad Gateway - Upstream server error
    res.Status(proteus.Status502)
    res.Send("Bad gateway - upstream server error")
})

server.Get("/maintenance", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // 503 Service Unavailable - Temporary unavailability
    res.Status(proteus.Status503)
    res.AddHeader("Retry-After", "3600") // Retry after 1 hour
    res.Send("Service temporarily unavailable for maintenance")
})

server.Get("/timeout", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // 504 Gateway Timeout - Upstream timeout
    res.Status(proteus.Status504)
    res.Send("Gateway timeout - upstream server did not respond")
})
```

### Error Handling Patterns

```go
// Global error handler middleware
func errorHandler(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
    defer func() {
        if r := recover(); r != nil {
            // Log the error
            fmt.Printf("Panic recovered: %v\n", r)

            // Send 500 Internal Server Error
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

// Database connection error simulation
server.Get("/api/data", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // Simulate database connection
    data, err := fetchFromDatabase()
    if err != nil {
        if err.Error() == "connection refused" {
            // 503 Service Unavailable - Database down
            res.Status(proteus.Status503)
            res.AddHeader("Retry-After", "300") // Retry after 5 minutes
            res.Send("Service temporarily unavailable - database maintenance")
        } else {
            // 500 Internal Server Error - Other database errors
            res.Status(proteus.Status500)
            res.Send("Database error occurred")
        }
        return
    }

    res.AddHeader("Content-Type", "application/json")
    res.Send(data)
})

func fetchFromDatabase() (string, error) {
    // Simulate random database errors
    switch time.Now().Second() % 10 {
    case 0:
        return "", fmt.Errorf("connection refused")
    case 1:
        return "", fmt.Errorf("query timeout")
    default:
        return `{"message": "Data from database"}`, nil
    }
}
```

## Complete Status Code Reference

```go
// All available status codes in Proteus
const (
    // 1xx Informational
    Status100 = 100 // Continue
    Status101 = 101 // Switching Protocols
    Status102 = 102 // Processing

    // 2xx Success
    Status200 = 200 // OK
    Status201 = 201 // Created
    Status202 = 202 // Accepted
    Status204 = 204 // No Content
    Status206 = 206 // Partial Content

    // 3xx Redirection
    Status301 = 301 // Moved Permanently
    Status302 = 302 // Found
    Status303 = 303 // See Other
    Status304 = 304 // Not Modified
    Status307 = 307 // Temporary Redirect
    Status308 = 308 // Permanent Redirect

    // 4xx Client Error
    Status400 = 400 // Bad Request
    Status401 = 401 // Unauthorized
    Status403 = 403 // Forbidden
    Status404 = 404 // Not Found
    Status405 = 405 // Method Not Allowed
    Status406 = 406 // Not Acceptable
    Status408 = 408 // Request Timeout
    Status409 = 409 // Conflict
    Status410 = 410 // Gone
    Status411 = 411 // Length Required
    Status413 = 413 // Payload Too Large
    Status415 = 415 // Unsupported Media Type
    Status422 = 422 // Unprocessable Entity
    Status429 = 429 // Too Many Requests

    // 5xx Server Error
    Status500 = 500 // Internal Server Error
    Status501 = 501 // Not Implemented
    Status502 = 502 // Bad Gateway
    Status503 = 503 // Service Unavailable
    Status504 = 504 // Gateway Timeout
)
```

## Best Practices

### When to Use Each Status Code

**2xx Success**
- `200 OK` - Standard successful response
- `201 Created` - Resource created (POST requests)
- `204 No Content` - Success with no response body (DELETE requests)

**3xx Redirection**
- `301 Moved Permanently` - URL permanently changed
- `302 Found` - Temporary redirect
- `304 Not Modified` - Resource hasn't changed (caching)

**4xx Client Errors**
- `400 Bad Request` - Invalid request data/format
- `401 Unauthorized` - Authentication required/failed
- `403 Forbidden` - Access denied (authenticated but insufficient permissions)
- `404 Not Found` - Resource doesn't exist
- `429 Too Many Requests` - Rate limiting

**5xx Server Errors**
- `500 Internal Server Error` - Generic server error
- `503 Service Unavailable` - Server overloaded or maintenance

### Status Code Guidelines

1. **Be Specific** - Use the most appropriate status code
2. **Be Consistent** - Use the same codes for similar situations
3. **Include Headers** - Add relevant headers (Location, Retry-After, etc.)
4. **Provide Context** - Include error messages in response body
5. **Log Errors** - Log 5xx errors for debugging
6. **Handle Gracefully** - Always handle error cases properly
