# File

## Overview

The `File` type in Proteus represents a static file that can be served by the web server. It provides information about the file including its path, content type, size, and modification time. Files are primarily used with the static file serving functionality to deliver assets like HTML pages, CSS stylesheets, JavaScript files, images, and other static content.

## Static File Serving

### Basic Static File Setup

```go
server := proteus.CreateServer()

// Serve static files from a directory
server.Static("/static", "/var/www/public")
server.Static("/assets", "/path/to/assets")
server.Static("/uploads", "/user/uploads")

// Multiple static directories
server.Static("/css", "/assets/stylesheets")
server.Static("/js", "/assets/javascript")
server.Static("/images", "/assets/images")
server.Static("/fonts", "/assets/fonts")
```

### File Path Mapping

```go
// URL path to filesystem mapping examples:
// server.Static("/static", "/var/www/public")

// Request: GET /static/css/style.css
// File: /var/www/public/css/style.css

// Request: GET /static/js/app.js
// File: /var/www/public/js/app.js

// Request: GET /static/images/logo.png
// File: /var/www/public/images/logo.png

// Request: GET /static/index.html
// File: /var/www/public/index.html
```

## File Properties

When Proteus serves a file, it automatically determines various properties:

### Content Type Detection

```go
// Proteus automatically detects content types based on file extensions:

// Text files
// .html, .htm -> text/html; charset=utf-8
// .css -> text/css; charset=utf-8
// .js -> application/javascript; charset=utf-8
// .json -> application/json; charset=utf-8
// .txt -> text/plain; charset=utf-8
// .xml -> application/xml; charset=utf-8

// Images
// .jpg, .jpeg -> image/jpeg
// .png -> image/png
// .gif -> image/gif
// .svg -> image/svg+xml
// .ico -> image/x-icon
// .webp -> image/webp

// Fonts
// .woff -> font/woff
// .woff2 -> font/woff2
// .ttf -> font/ttf
// .eot -> application/vnd.ms-fontobject

// Documents
// .pdf -> application/pdf
// .doc -> application/msword
// .docx -> application/vnd.openxmlformats-officedocument.wordprocessingml.document

// Archives
// .zip -> application/zip
// .tar -> application/x-tar
// .gz -> application/gzip

// Default for unknown extensions
// * -> application/octet-stream
```

### File Headers

When serving static files, Proteus automatically sets appropriate HTTP headers:

```go
// Example headers set for static files:

// Content-Type header based on file extension
// Content-Length header with file size
// Last-Modified header with file modification time
// ETag header for caching (based on file info)

// For an image file request:
// Content-Type: image/jpeg
// Content-Length: 245760
// Last-Modified: Wed, 21 Oct 2023 07:28:00 GMT
// ETag: "3c000-5f8d2f00-a1b2c3"

// For a CSS file request:
// Content-Type: text/css; charset=utf-8
// Content-Length: 12543
// Last-Modified: Thu, 22 Oct 2023 14:30:15 GMT
// ETag: "30ff-5f8e8107-d4e5f6"
```

## Custom File Handling

### Manual File Serving

```go
server.Get("/download/:filename", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    filename, _ := req.Segments.Get("filename")
    filePath := "/var/downloads/" + filename[0]

    // Check if file exists
    fileInfo, err := os.Stat(filePath)
    if err != nil {
        res.Status(proteus.Status404)
        res.Send("File not found")
        return
    }

    // Read file content
    content, err := ioutil.ReadFile(filePath)
    if err != nil {
        res.Status(proteus.Status500)
        res.Send("Error reading file")
        return
    }

    // Detect content type from extension
    contentType := getContentType(filePath)

    // Set file headers
    res.AddHeader("Content-Type", contentType)
    res.AddHeader("Content-Length", fmt.Sprintf("%d", len(content)))
    res.AddHeader("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename[0]))
    res.AddHeader("Last-Modified", fileInfo.ModTime().Format(http.TimeFormat))

    res.Send(string(content))
})

func getContentType(filePath string) string {
    ext := strings.ToLower(filepath.Ext(filePath))

    switch ext {
    case ".html", ".htm":
        return "text/html; charset=utf-8"
    case ".css":
        return "text/css; charset=utf-8"
    case ".js":
        return "application/javascript; charset=utf-8"
    case ".json":
        return "application/json; charset=utf-8"
    case ".png":
        return "image/png"
    case ".jpg", ".jpeg":
        return "image/jpeg"
    case ".gif":
        return "image/gif"
    case ".pdf":
        return "application/pdf"
    case ".txt":
        return "text/plain; charset=utf-8"
    default:
        return "application/octet-stream"
    }
}
```

### File Upload Handling

```go
server.Post("/upload", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // For file uploads, you'd typically handle multipart/form-data
    contentType, _ := req.Headers.Get("Content-Type")

    if !strings.Contains(contentType, "multipart/form-data") {
        res.Status(proteus.Status400)
        res.Send("Multipart form data required for file upload")
        return
    }

    // In a real implementation, you would parse multipart data
    // For now, we'll simulate file upload handling

    rawBody := req.Body.(string)

    // Extract filename from Content-Disposition header in multipart data
    // This is simplified - real multipart parsing is more complex

    uploadDir := "/var/uploads/"
    filename := generateUniqueFilename()
    filePath := uploadDir + filename

    // Save uploaded content (simplified)
    err := ioutil.WriteFile(filePath, []byte(rawBody), 0644)
    if err != nil {
        res.Status(proteus.Status500)
        res.Send("Error saving file")
        return
    }

    // Return file info
    res.AddHeader("Content-Type", "application/json")
    res.Status(proteus.Status201)
    res.Send(fmt.Sprintf(`{
        "message": "File uploaded successfully",
        "filename": "%s",
        "path": "%s",
        "size": %d
    }`, filename, filePath, len(rawBody)))
})

func generateUniqueFilename() string {
    return fmt.Sprintf("upload_%d", time.Now().UnixNano())
}
```

## File Security

### Path Traversal Prevention

```go
server.Get("/secure-files/*", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    pathParam, exists := req.Segments.Get("*")
    if !exists || len(pathParam) == 0 {
        res.Status(proteus.Status400)
        res.Send("File path required")
        return
    }

    requestedPath := pathParam[0]

    // Security: Prevent path traversal attacks
    if strings.Contains(requestedPath, "..") {
        res.Status(proteus.Status400)
        res.Send("Invalid file path")
        return
    }

    if strings.HasPrefix(requestedPath, "/") {
        res.Status(proteus.Status400)
        res.Send("Absolute paths not allowed")
        return
    }

    // Clean and validate the path
    cleanPath := filepath.Clean(requestedPath)
    if cleanPath != requestedPath {
        res.Status(proteus.Status400)
        res.Send("Invalid file path")
        return
    }

    // Construct safe file path
    safeDir := "/var/secure-files/"
    fullPath := filepath.Join(safeDir, cleanPath)

    // Verify the resolved path is still within safe directory
    if !strings.HasPrefix(fullPath, safeDir) {
        res.Status(proteus.Status403)
        res.Send("Access denied")
        return
    }

    // Serve the file
    content, err := ioutil.ReadFile(fullPath)
    if err != nil {
        res.Status(proteus.Status404)
        res.Send("File not found")
        return
    }

    res.AddHeader("Content-Type", getContentType(fullPath))
    res.Send(string(content))
})
```

### File Access Control

```go
// Authentication middleware for protected files
fileAuthMiddleware := func(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
    // Check authentication
    token, exists := req.Headers.Get("Authorization")
    if !exists || !strings.HasPrefix(token, "Bearer ") {
        res.Status(proteus.Status401)
        res.Send("Authentication required")
        return
    }

    // Validate token and get user
    user := validateToken(token[7:])
    if user == nil {
        res.Status(proteus.Status401)
        res.Send("Invalid token")
        return
    }

    req.Locals["user"] = user
    next()
}

server.Get("/private-files/*", fileAuthMiddleware,
    func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
        user := req.Locals["user"].(User)
        pathParam, _ := req.Segments.Get("*")
        requestedFile := pathParam[0]

        // Check if user has access to this file
        if !userHasFileAccess(user, requestedFile) {
            res.Status(proteus.Status403)
            res.Send("Access denied to this file")
            return
        }

        // Serve the file
        filePath := "/var/private-files/" + requestedFile
        content, err := ioutil.ReadFile(filePath)
        if err != nil {
            res.Status(proteus.Status404)
            res.Send("File not found")
            return
        }

        res.AddHeader("Content-Type", getContentType(filePath))
        res.Send(string(content))
    })

func userHasFileAccess(user User, filename string) bool {
    // Check file permissions based on user role/ownership
    return user.IsAdmin || strings.HasPrefix(filename, "user_"+user.ID+"_")
}

type User struct {
    ID      string
    IsAdmin bool
}

func validateToken(token string) *User {
    // Simplified token validation
    if token == "valid-user-token" {
        return &User{ID: "123", IsAdmin: false}
    }
    if token == "admin-token" {
        return &User{ID: "admin", IsAdmin: true}
    }
    return nil
}
```

## Caching and Performance

### File Caching Headers

```go
server.Use("/assets/*", func(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
    // Set caching headers for static assets
    ext := filepath.Ext(req.ResourcePath)

    switch strings.ToLower(ext) {
    case ".css", ".js", ".png", ".jpg", ".jpeg", ".gif", ".woff", ".woff2":
        // Long cache for static assets (1 year)
        res.AddHeader("Cache-Control", "public, max-age=31536000, immutable")
        res.AddHeader("Expires", time.Now().AddDate(1, 0, 0).Format(http.TimeFormat))

    case ".html", ".htm":
        // Short cache for HTML (1 hour)
        res.AddHeader("Cache-Control", "public, max-age=3600")

    case ".json", ".xml":
        // No cache for API responses
        res.AddHeader("Cache-Control", "no-cache, no-store, must-revalidate")
        res.AddHeader("Pragma", "no-cache")
        res.AddHeader("Expires", "0")

    default:
        // Default caching (1 day)
        res.AddHeader("Cache-Control", "public, max-age=86400")
    }

    next()
})
```

### Conditional Requests

```go
server.Get("/api/file/:id", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    fileID, _ := req.Segments.Get("id")
    filePath := "/var/files/" + fileID[0]

    // Get file info
    fileInfo, err := os.Stat(filePath)
    if err != nil {
        res.Status(proteus.Status404)
        res.Send("File not found")
        return
    }

    // Generate ETag based on file info
    etag := fmt.Sprintf("\"%x-%x\"", fileInfo.Size(), fileInfo.ModTime().Unix())
    lastModified := fileInfo.ModTime().Format(http.TimeFormat)

    // Check If-None-Match header (ETag-based caching)
    if clientETag, exists := req.Headers.Get("If-None-Match"); exists {
        if clientETag == etag {
            res.Status(proteus.Status304) // Not Modified
            res.AddHeader("ETag", etag)
            res.AddHeader("Last-Modified", lastModified)
            res.Send("")
            return
        }
    }

    // Check If-Modified-Since header (time-based caching)
    if ifModSince, exists := req.Headers.Get("If-Modified-Since"); exists {
        if clientTime, err := http.ParseTime(ifModSince); err == nil {
            if fileInfo.ModTime().Before(clientTime.Add(time.Second)) {
                res.Status(proteus.Status304) // Not Modified
                res.AddHeader("ETag", etag)
                res.AddHeader("Last-Modified", lastModified)
                res.Send("")
                return
            }
        }
    }

    // File has been modified, send full content
    content, err := ioutil.ReadFile(filePath)
    if err != nil {
        res.Status(proteus.Status500)
        res.Send("Error reading file")
        return
    }

    res.AddHeader("Content-Type", getContentType(filePath))
    res.AddHeader("Content-Length", fmt.Sprintf("%d", len(content)))
    res.AddHeader("ETag", etag)
    res.AddHeader("Last-Modified", lastModified)
    res.AddHeader("Cache-Control", "public, max-age=3600")

    res.Send(string(content))
})
```

## Key Features

### Static File Serving
- Automatic content type detection based on file extension
- Support for nested directory structures
- Security features to prevent path traversal
- Configurable cache headers for performance

### File Properties
- Content type detection for proper MIME types
- File size and modification time tracking
- ETag generation for efficient caching
- Support for conditional requests (304 Not Modified)

### Security Features
- Path traversal prevention (no ".." in paths)
- Access control through middleware
- File permission validation
- Safe path construction and validation

### Performance Optimization
- HTTP caching headers (Cache-Control, ETag, Last-Modified)
- Conditional request handling (If-None-Match, If-Modified-Since)
- Content compression support
- Long-term caching for static assets

### Use Cases
- Serving web application assets (HTML, CSS, JS)
- Image and media file delivery
- Document downloads and file sharing
- User-uploaded content serving
- API file endpoints with access control
