# Headers

## Overview

The `Headers` collection in Proteus provides access to HTTP headers from incoming requests and allows you to set headers on outgoing responses. Headers are key-value pairs that carry metadata about the HTTP request or response, such as content type, authentication tokens, caching directives, and custom application data.

## Request Headers

### Basic Header Access

```go
server.Get("/info", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // Get a specific header
    userAgent, exists := req.Headers.Get("User-Agent")
    if exists {
        res.Send("Your browser: " + userAgent)
    } else {
        res.Send("User-Agent header not found")
    }

    // Check if header exists without getting value
    if req.Headers.Has("Accept-Language") {
        language, _ := req.Headers.Get("Accept-Language")
        res.Send("Preferred language: " + language)
    }

    // Get host header
    host, _ := req.Headers.Get("Host")
    res.Send("Request host: " + host)
})
```

### Common Request Headers

```go
server.Get("/analyze", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // Content-Type (for requests with body)
    contentType, hasContentType := req.Headers.Get("Content-Type")

    // User-Agent (browser/client information)
    userAgent, _ := req.Headers.Get("User-Agent")

    // Accept (what content types client accepts)
    accept, _ := req.Headers.Get("Accept")

    // Accept-Language (preferred languages)
    acceptLang, _ := req.Headers.Get("Accept-Language")

    // Accept-Encoding (supported compression)
    acceptEncoding, _ := req.Headers.Get("Accept-Encoding")

    // Host (server hostname)
    host, _ := req.Headers.Get("Host")

    // Connection (connection preferences)
    connection, _ := req.Headers.Get("Connection")

    // Referer (previous page)
    referer, hasReferer := req.Headers.Get("Referer")

    analysis := fmt.Sprintf(`Request Analysis:
- Content-Type: %s (present: %t)
- User-Agent: %s
- Accept: %s
- Accept-Language: %s
- Accept-Encoding: %s
- Host: %s
- Connection: %s
- Referer: %s (present: %t)`,
        contentType, hasContentType, userAgent, accept, acceptLang,
        acceptEncoding, host, connection, referer, hasReferer)

    res.Send(analysis)
})
```

## Response Headers

### Setting Response Headers

```go
server.Get("/api/data", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // Set content type
    res.AddHeader("Content-Type", "application/json; charset=utf-8")

    // Set caching headers
    res.AddHeader("Cache-Control", "public, max-age=3600")
    res.AddHeader("ETag", "\"abc123\"")
    res.AddHeader("Last-Modified", time.Now().Format(http.TimeFormat))

    // Set custom headers
    res.AddHeader("X-API-Version", "1.0")
    res.AddHeader("X-Rate-Limit-Remaining", "99")
    res.AddHeader("X-Request-ID", generateRequestID())

    data := `{"message": "Hello World", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`
    res.Send(data)
})
```

### Security Headers

```go
// Security headers middleware
server.Use(func(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
    // Prevent clickjacking
    res.AddHeader("X-Frame-Options", "DENY")

    // XSS protection
    res.AddHeader("X-XSS-Protection", "1; mode=block")

    // Content type sniffing protection
    res.AddHeader("X-Content-Type-Options", "nosniff")

    // HSTS (HTTPS only)
    res.AddHeader("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

    // Content Security Policy
    res.AddHeader("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'")

    // Referrer Policy
    res.AddHeader("Referrer-Policy", "strict-origin-when-cross-origin")

    // Permissions Policy
    res.AddHeader("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

    next()
})
```

## Content Negotiation

### Content Type Detection

```go
server.Post("/upload", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    contentType, exists := req.Headers.Get("Content-Type")

    if !exists {
        res.Status(proteus.Status400)
        res.Send("Content-Type header required")
        return
    }

    switch {
    case strings.Contains(contentType, "application/json"):
        // Handle JSON
        res.Send("Processing JSON data")

    case strings.Contains(contentType, "application/x-www-form-urlencoded"):
        // Handle form data
        res.Send("Processing form data")

    case strings.Contains(contentType, "multipart/form-data"):
        // Handle file upload
        res.Send("Processing file upload")

    case strings.Contains(contentType, "text/plain"):
        // Handle plain text
        res.Send("Processing plain text")

    case strings.Contains(contentType, "application/xml"):
        // Handle XML
        res.Send("Processing XML data")

    default:
        res.Status(proteus.Status415)
        res.Send("Unsupported Media Type: " + contentType)
    }
})
```

### Accept Header Processing

```go
server.Get("/data", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    accept, _ := req.Headers.Get("Accept")

    // Sample data
    data := map[string]interface{}{
        "id":   123,
        "name": "John Doe",
        "email": "john@example.com",
    }

    switch {
    case strings.Contains(accept, "application/json"):
        res.AddHeader("Content-Type", "application/json")
        jsonData := fmt.Sprintf(`{
            "id": %d,
            "name": "%s",
            "email": "%s"
        }`, data["id"], data["name"], data["email"])
        res.Send(jsonData)

    case strings.Contains(accept, "application/xml"):
        res.AddHeader("Content-Type", "application/xml")
        xmlData := fmt.Sprintf(`<?xml version="1.0"?>
        <user>
            <id>%d</id>
            <name>%s</name>
            <email>%s</email>
        </user>`, data["id"], data["name"], data["email"])
        res.Send(xmlData)

    case strings.Contains(accept, "text/html"):
        res.AddHeader("Content-Type", "text/html")
        htmlData := fmt.Sprintf(`<!DOCTYPE html>
        <html>
        <body>
            <h1>User Profile</h1>
            <p>ID: %d</p>
            <p>Name: %s</p>
            <p>Email: %s</p>
        </body>
        </html>`, data["id"], data["name"], data["email"])
        res.Send(htmlData)

    case strings.Contains(accept, "text/plain"):
        res.AddHeader("Content-Type", "text/plain")
        textData := fmt.Sprintf("User: %s (ID: %d) - %s",
            data["name"], data["id"], data["email"])
        res.Send(textData)

    default:
        // Default to JSON
        res.AddHeader("Content-Type", "application/json")
        jsonData := fmt.Sprintf(`{"id": %d, "name": "%s", "email": "%s"}`,
            data["id"], data["name"], data["email"])
        res.Send(jsonData)
    }
})
```

## Custom Headers

### Application-Specific Headers

```go
// Request tracking middleware
server.Use(func(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
    // Generate request ID
    requestID := generateUUID()

    // Check if client provided request ID
    if clientID, exists := req.Headers.Get("X-Request-ID"); exists {
        requestID = clientID
    }

    // Store in locals and add to response
    req.Locals["requestID"] = requestID
    res.AddHeader("X-Request-ID", requestID)

    // Add processing start time
    req.Locals["startTime"] = time.Now()

    next()

    // Add processing time to response
    duration := time.Since(req.Locals["startTime"].(time.Time))
    res.AddHeader("X-Processing-Time", duration.String())
})

// Rate limiting headers
server.Use(func(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
    clientIP := getClientIP(req)

    // Get rate limit info (simplified)
    remaining := getRemainingRequests(clientIP)
    resetTime := getResetTime(clientIP)

    // Add rate limit headers
    res.AddHeader("X-RateLimit-Limit", "100")
    res.AddHeader("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
    res.AddHeader("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))

    if remaining <= 0 {
        res.Status(proteus.Status429)
        res.AddHeader("Retry-After", fmt.Sprintf("%d", int(time.Until(resetTime).Seconds())))
        res.Send("Rate limit exceeded")
        return
    }

    next()
})
```

## Method Reference

### Header Collection Methods
- `Get(name)` (string, bool) - Get header value and existence check
- `Has(name)` bool - Check if header exists
- `Set(name, value)` - Set header (mainly for testing)
- `Delete(name)` - Remove header (mainly for testing)

### Common Request Headers
- `Authorization` - Authentication credentials
- `Content-Type` - Body content type
- `Accept` - Acceptable response formats
- `User-Agent` - Client information
- `Host` - Target hostname
- `Referer` - Previous page URL

### Common Response Headers
- `Content-Type` - Response content type
- `Content-Length` - Response body size
- `Cache-Control` - Caching directives
- `Set-Cookie` - Cookie instructions
- `Location` - Redirect target
- `Access-Control-*` - CORS headers
