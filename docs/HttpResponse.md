# HttpResponse

## Overview

The `HttpResponse` object represents the HTTP response that your Proteus server sends back to the client. It provides methods for setting headers, status codes, and sending response content. The response object is automatically passed as the second parameter to all route handlers.

## Basic Response Operations

### Sending Text Content

```go
server.Get("/", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // Send simple text response
    res.Send("Hello, World!")
})

server.Get("/info", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // Send formatted text
    message := fmt.Sprintf("Welcome, user! Current time: %s", time.Now().Format(time.RFC3339))
    res.Send(message)
})

server.Get("/html", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // Send HTML content
    html := `
    <!DOCTYPE html>
    <html>
    <head><title>Proteus App</title></head>
    <body><h1>Welcome to Proteus!</h1></body>
    </html>`

    res.AddHeader("Content-Type", "text/html; charset=utf-8")
    res.Send(html)
})
```

### Setting Status Codes

```go
server.Get("/success", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    res.Status(proteus.Status200)  // OK
    res.Send("Operation successful")
})

server.Post("/api/users", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // Create user logic...

    res.Status(proteus.Status201)  // Created
    res.Send("User created successfully")
})

server.Get("/not-found", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    res.Status(proteus.Status404)  // Not Found
    res.Send("Resource not found")
})

server.Get("/error", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    res.Status(proteus.Status500)  // Internal Server Error
    res.Send("Something went wrong")
})

server.Delete("/api/users/:id", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    userID, _ := req.Segments.Get("id")

    // Delete user logic...

    res.Status(proteus.Status204)  // No Content
    res.Send("")  // Empty body for 204
})
```

## Response Headers

### Setting Individual Headers

```go
server.Get("/api/data", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // Set content type
    res.AddHeader("Content-Type", "application/json")

    // Set cache control
    res.AddHeader("Cache-Control", "no-cache, no-store, must-revalidate")

    // Set custom headers
    res.AddHeader("X-API-Version", "1.0")
    res.AddHeader("X-Request-ID", generateRequestID())

    data := `{"message": "Hello from API", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`
    res.Send(data)
})

server.Get("/download", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // Set headers for file download
    res.AddHeader("Content-Type", "application/octet-stream")
    res.AddHeader("Content-Disposition", "attachment; filename=\"data.txt\"")
    res.AddHeader("Content-Length", "12")

    res.Send("Sample data")
})
```

### Common Header Patterns

```go
// CORS headers
server.Use(func(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
    res.AddHeader("Access-Control-Allow-Origin", "*")
    res.AddHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    res.AddHeader("Access-Control-Allow-Headers", "Content-Type, Authorization")
    res.AddHeader("Access-Control-Max-Age", "3600")

    if req.Method == "OPTIONS" {
        res.Status(proteus.Status200)
        res.Send("")
        return
    }

    next()
})

// Security headers
server.Use(func(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
    res.AddHeader("X-Frame-Options", "DENY")
    res.AddHeader("X-XSS-Protection", "1; mode=block")
    res.AddHeader("X-Content-Type-Options", "nosniff")
    res.AddHeader("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
    res.AddHeader("Content-Security-Policy", "default-src 'self'")

    next()
})

// API response headers
server.Get("/api/*", func(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
    res.AddHeader("Content-Type", "application/json; charset=utf-8")
    res.AddHeader("X-Powered-By", "Proteus")
    res.AddHeader("X-API-Version", "v1")

    next()
})
```

## JSON Responses

### Manual JSON Construction

```go
server.Get("/api/user/:id", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    userID, _ := req.Segments.Get("id")

    // Simulate user data
    user := map[string]interface{}{
        "id":    userID[0],
        "name":  "John Doe",
        "email": "john@example.com",
        "active": true,
    }

    res.AddHeader("Content-Type", "application/json")
    res.Send(fmt.Sprintf(`{
        "id": "%s",
        "name": "%s",
        "email": "%s",
        "active": %t,
        "timestamp": "%s"
    }`, user["id"], user["name"], user["email"], user["active"], time.Now().Format(time.RFC3339)))
})

server.Get("/api/users", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // Simulate user list
    users := []map[string]interface{}{
        {"id": "1", "name": "John Doe", "email": "john@example.com"},
        {"id": "2", "name": "Jane Smith", "email": "jane@example.com"},
    }

    res.AddHeader("Content-Type", "application/json")

    jsonResponse := `{
        "users": [`

    for i, user := range users {
        if i > 0 {
            jsonResponse += ","
        }
        jsonResponse += fmt.Sprintf(`{
            "id": "%s",
            "name": "%s",
            "email": "%s"
        }`, user["id"], user["name"], user["email"])
    }

    jsonResponse += `],
        "count": ` + fmt.Sprintf("%d", len(users)) + `,
        "timestamp": "` + time.Now().Format(time.RFC3339) + `"
    }`

    res.Send(jsonResponse)
})
```

### Structured JSON Responses

```go
// Success response helper
func sendSuccess(res *proteus.HttpResponse, message string, data interface{}) {
    res.AddHeader("Content-Type", "application/json")
    res.Status(proteus.Status200)

    response := fmt.Sprintf(`{
        "success": true,
        "message": "%s",
        "data": %+v,
        "timestamp": "%s"
    }`, message, data, time.Now().Format(time.RFC3339))

    res.Send(response)
}

// Error response helper
func sendError(res *proteus.HttpResponse, statusCode proteus.StatusCode, message string) {
    res.AddHeader("Content-Type", "application/json")
    res.Status(statusCode)

    response := fmt.Sprintf(`{
        "success": false,
        "error": "%s",
        "timestamp": "%s"
    }`, message, time.Now().Format(time.RFC3339))

    res.Send(response)
}

// Usage in handlers
server.Get("/api/posts", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    posts := []string{"Post 1", "Post 2", "Post 3"}
    sendSuccess(res, "Posts retrieved successfully", posts)
})

server.Get("/api/posts/:id", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    postID, exists := req.Segments.Get("id")
    if !exists || len(postID) == 0 {
        sendError(res, proteus.Status400, "Post ID is required")
        return
    }

    // Simulate post lookup
    if postID[0] == "999" {
        sendError(res, proteus.Status404, "Post not found")
        return
    }

    post := map[string]interface{}{
        "id": postID[0],
        "title": "Sample Post",
        "content": "This is a sample post content",
    }

    sendSuccess(res, "Post retrieved successfully", post)
})
```

## File Responses

### Serving Files

```go
server.Get("/files/:filename", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    filename, _ := req.Segments.Get("filename")
    filepath := "/var/www/uploads/" + filename[0]

    // Read file content
    content, err := ioutil.ReadFile(filepath)
    if err != nil {
        res.Status(proteus.Status404)
        res.Send("File not found")
        return
    }

    // Detect content type based on file extension
    ext := strings.ToLower(filepath[strings.LastIndex(filepath, "."):])

    var contentType string
    switch ext {
    case ".html", ".htm":
        contentType = "text/html; charset=utf-8"
    case ".css":
        contentType = "text/css; charset=utf-8"
    case ".js":
        contentType = "application/javascript; charset=utf-8"
    case ".json":
        contentType = "application/json; charset=utf-8"
    case ".png":
        contentType = "image/png"
    case ".jpg", ".jpeg":
        contentType = "image/jpeg"
    case ".gif":
        contentType = "image/gif"
    case ".pdf":
        contentType = "application/pdf"
    default:
        contentType = "application/octet-stream"
    }

    res.AddHeader("Content-Type", contentType)
    res.AddHeader("Content-Length", fmt.Sprintf("%d", len(content)))
    res.Send(string(content))
})

server.Get("/download/:filename", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    filename, _ := req.Segments.Get("filename")

    // Force download regardless of content type
    res.AddHeader("Content-Type", "application/octet-stream")
    res.AddHeader("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename[0]))

    // Serve file content...
    res.Send("File content here")
})
```

### Image and Media Responses

```go
server.Get("/images/:id", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    imageID, _ := req.Segments.Get("id")
    imagePath := "/var/images/" + imageID[0] + ".jpg"

    // Check if file exists
    if _, err := os.Stat(imagePath); os.IsNotExist(err) {
        res.Status(proteus.Status404)
        res.Send("Image not found")
        return
    }

    // Read image file
    imageData, err := ioutil.ReadFile(imagePath)
    if err != nil {
        res.Status(proteus.Status500)
        res.Send("Error reading image")
        return
    }

    // Set appropriate headers
    res.AddHeader("Content-Type", "image/jpeg")
    res.AddHeader("Content-Length", fmt.Sprintf("%d", len(imageData)))
    res.AddHeader("Cache-Control", "public, max-age=31536000") // Cache for 1 year
    res.AddHeader("ETag", fmt.Sprintf("\"%s\"", generateETag(imageData)))

    res.Send(string(imageData))
})
```

## Streaming and Large Responses

```go
server.Get("/large-data", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    res.AddHeader("Content-Type", "text/plain")
    res.AddHeader("Transfer-Encoding", "chunked")

    // Send data in chunks
    for i := 0; i < 1000; i++ {
        chunk := fmt.Sprintf("Data chunk %d\n", i)
        res.Send(chunk)

        // In a real implementation, you might flush here
        time.Sleep(10 * time.Millisecond) // Simulate processing time
    }
})

server.Get("/csv-export", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    res.AddHeader("Content-Type", "text/csv; charset=utf-8")
    res.AddHeader("Content-Disposition", "attachment; filename=\"export.csv\"")

    // Send CSV header
    res.Send("ID,Name,Email,Created\n")

    // Send CSV data rows
    users := []map[string]string{
        {"id": "1", "name": "John Doe", "email": "john@example.com"},
        {"id": "2", "name": "Jane Smith", "email": "jane@example.com"},
    }

    for _, user := range users {
        row := fmt.Sprintf("%s,\"%s\",\"%s\",\"%s\"\n",
            user["id"], user["name"], user["email"], time.Now().Format("2006-01-02"))
        res.Send(row)
    }
})
```

## Response Middleware

```go
// Response timing middleware
server.Use(func(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
    start := time.Now()

    next()

    duration := time.Since(start)
    res.AddHeader("X-Response-Time", duration.String())
})

// Response compression simulation
server.Use(func(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
    acceptEncoding, _ := req.Headers.Get("Accept-Encoding")

    if strings.Contains(acceptEncoding, "gzip") {
        res.AddHeader("Content-Encoding", "gzip")
    }

    next()
})

// API versioning middleware
server.Use("/api/v1/*", func(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
    res.AddHeader("X-API-Version", "1.0")
    res.AddHeader("X-Deprecation-Warning", "API v1 will be deprecated on 2024-12-31")

    next()
})
```

## Error Handling

```go
// Custom error responses
server.Get("/api/divide/:a/:b", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    aStr, _ := req.Segments.Get("a")
    bStr, _ := req.Segments.Get("b")

    a, errA := strconv.Atoi(aStr[0])
    b, errB := strconv.Atoi(bStr[0])

    if errA != nil || errB != nil {
        res.Status(proteus.Status400)
        res.AddHeader("Content-Type", "application/json")
        res.Send(`{
            "error": "Invalid input",
            "message": "Both parameters must be integers",
            "code": "INVALID_INPUT"
        }`)
        return
    }

    if b == 0 {
        res.Status(proteus.Status400)
        res.AddHeader("Content-Type", "application/json")
        res.Send(`{
            "error": "Division by zero",
            "message": "Cannot divide by zero",
            "code": "DIVISION_BY_ZERO"
        }`)
        return
    }

    result := float64(a) / float64(b)
    res.AddHeader("Content-Type", "application/json")
    res.Send(fmt.Sprintf(`{
        "result": %.2f,
        "operation": "%d ÷ %d"
    }`, result, a, b))
})

// Global error handler
server.Use(func(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
    defer func() {
        if r := recover(); r != nil {
            res.Status(proteus.Status500)
            res.AddHeader("Content-Type", "application/json")
            res.Send(`{
                "error": "Internal Server Error",
                "message": "An unexpected error occurred",
                "code": "INTERNAL_ERROR"
            }`)
        }
    }()

    next()
})
```

## Method Reference

### Core Methods
- `Send(content)` - Send response content to client
- `Status(statusCode)` - Set HTTP status code
- `AddHeader(name, value)` - Add HTTP response header

### Status Code Examples
- `proteus.Status200` - OK (200)
- `proteus.Status201` - Created (201)
- `proteus.Status204` - No Content (204)
- `proteus.Status400` - Bad Request (400)
- `proteus.Status401` - Unauthorized (401)
- `proteus.Status404` - Not Found (404)
- `proteus.Status500` - Internal Server Error (500)

### Common Content Types
- `text/plain` - Plain text
- `text/html; charset=utf-8` - HTML content
- `application/json; charset=utf-8` - JSON data
- `application/octet-stream` - Binary files
- `image/jpeg`, `image/png` - Images
- `text/css` - CSS stylesheets
- `application/javascript` - JavaScript files
