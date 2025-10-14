# Params

## Overview

The `Params` collection in Proteus provides access to URL parameters from two sources: query string parameters (after the `?` in the URL) and path segments (route parameters defined with `:` in route patterns). Both collections use the same interface but serve different purposes in handling HTTP requests.

## Query String Parameters

Query parameters are key-value pairs that appear after the `?` in a URL, separated by `&` characters.

### Basic Query Parameter Access

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

    // Check parameter existence
    if req.Params.Has("debug") {
        res.Send("Debug mode enabled")
    }

    // Multiple values for same parameter
    tags, hasTags := req.Params.Get("tags")
    if hasTags {
        res.Send(fmt.Sprintf("Tags: %v", tags))
    }
})

// Example URLs:
// /search?q=golang                    -> query = ["golang"]
// /search?q=golang&debug=true         -> query = ["golang"], debug exists
// /search?tags=web&tags=framework     -> tags = ["web", "framework"]
// /search                             -> no parameters
```

### Common Query Parameter Patterns

```go
// Pagination parameters
server.Get("/posts", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // Page parameter with default
    page := 1
    if pageParam, exists := req.Params.Get("page"); exists && len(pageParam) > 0 {
        if p, err := strconv.Atoi(pageParam[0]); err == nil && p > 0 {
            page = p
        }
    }

    // Limit parameter with default and validation
    limit := 10
    if limitParam, exists := req.Params.Get("limit"); exists && len(limitParam) > 0 {
        if l, err := strconv.Atoi(limitParam[0]); err == nil && l > 0 && l <= 100 {
            limit = l
        }
    }

    // Sort parameter
    sortBy := "created_at"
    if sortParam, exists := req.Params.Get("sort"); exists && len(sortParam) > 0 {
        validSorts := []string{"created_at", "updated_at", "title", "author"}
        for _, valid := range validSorts {
            if sortParam[0] == valid {
                sortBy = sortParam[0]
                break
            }
        }
    }

    // Sort order
    order := "desc"
    if orderParam, exists := req.Params.Get("order"); exists && len(orderParam) > 0 {
        if orderParam[0] == "asc" || orderParam[0] == "desc" {
            order = orderParam[0]
        }
    }

    res.Send(fmt.Sprintf("Posts page %d, limit %d, sorted by %s %s",
        page, limit, sortBy, order))
})

// Example: /posts?page=2&limit=20&sort=title&order=asc
```

### Filtering and Search Parameters

```go
server.Get("/products", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // Search query
    search := ""
    if searchParam, exists := req.Params.Get("search"); exists && len(searchParam) > 0 {
        search = searchParam[0]
    }

    // Category filter (multiple allowed)
    categories, hasCategories := req.Params.Get("category")

    // Price range
    minPrice := 0.0
    maxPrice := 0.0

    if minParam, exists := req.Params.Get("min_price"); exists && len(minParam) > 0 {
        if price, err := strconv.ParseFloat(minParam[0], 64); err == nil && price >= 0 {
            minPrice = price
        }
    }

    if maxParam, exists := req.Params.Get("max_price"); exists && len(maxParam) > 0 {
        if price, err := strconv.ParseFloat(maxParam[0], 64); err == nil && price >= 0 {
            maxPrice = price
        }
    }

    // Boolean flags
    inStock := false
    if req.Params.Has("in_stock") {
        inStock = true
    }

    onSale := false
    if saleParam, exists := req.Params.Get("on_sale"); exists && len(saleParam) > 0 {
        onSale = saleParam[0] == "true" || saleParam[0] == "1"
    }

    // Build filter description
    filters := []string{}
    if search != "" {
        filters = append(filters, "search: "+search)
    }
    if hasCategories {
        filters = append(filters, "categories: "+strings.Join(categories, ", "))
    }
    if minPrice > 0 {
        filters = append(filters, fmt.Sprintf("min_price: %.2f", minPrice))
    }
    if maxPrice > 0 {
        filters = append(filters, fmt.Sprintf("max_price: %.2f", maxPrice))
    }
    if inStock {
        filters = append(filters, "in_stock: true")
    }
    if onSale {
        filters = append(filters, "on_sale: true")
    }

    if len(filters) > 0 {
        res.Send("Filtering products with: " + strings.Join(filters, ", "))
    } else {
        res.Send("No filters applied")
    }
})

// Example: /products?search=laptop&category=electronics&category=computers&min_price=500&max_price=2000&in_stock&on_sale=true
```

## Path Parameters (Route Segments)

Path parameters are defined in route patterns using `:paramName` syntax and extracted from the actual URL path.

### Basic Path Parameters

```go
// Single parameter
server.Get("/users/:id", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    userID, exists := req.Segments.Get("id")
    if exists && len(userID) > 0 {
        id := userID[0]
        res.Send("User ID: " + id)
    } else {
        res.Status(proteus.Status400)
        res.Send("Invalid user ID")
    }
})

// Multiple parameters
server.Get("/users/:userID/posts/:postID", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    userID, _ := req.Segments.Get("userID")
    postID, _ := req.Segments.Get("postID")

    res.Send(fmt.Sprintf("User %s, Post %s", userID[0], postID[0]))
})

// Nested resource parameters
server.Get("/api/:version/users/:userID/posts/:postID/comments/:commentID",
    func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
        version, _ := req.Segments.Get("version")
        userID, _ := req.Segments.Get("userID")
        postID, _ := req.Segments.Get("postID")
        commentID, _ := req.Segments.Get("commentID")

        response := fmt.Sprintf("API %s: User %s, Post %s, Comment %s",
            version[0], userID[0], postID[0], commentID[0])
        res.Send(response)
    })

// Examples:
// /users/123                                    -> userID = "123"
// /users/123/posts/456                         -> userID = "123", postID = "456"
// /api/v2/users/123/posts/456/comments/789     -> version = "v2", userID = "123", etc.
```

### Parameter Validation

```go
server.Get("/users/:id", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    userIDParam, exists := req.Segments.Get("id")
    if !exists || len(userIDParam) == 0 {
        res.Status(proteus.Status400)
        res.Send("User ID is required")
        return
    }

    // Validate ID is numeric
    userID, err := strconv.Atoi(userIDParam[0])
    if err != nil {
        res.Status(proteus.Status400)
        res.Send("User ID must be a number")
        return
    }

    // Validate ID range
    if userID <= 0 || userID > 999999 {
        res.Status(proteus.Status400)
        res.Send("User ID must be between 1 and 999999")
        return
    }

    res.Send(fmt.Sprintf("Valid user ID: %d", userID))
})

server.Get("/posts/:slug", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    slugParam, _ := req.Segments.Get("slug")
    slug := slugParam[0]

    // Validate slug format (lowercase letters, numbers, hyphens)
    if !isValidSlug(slug) {
        res.Status(proteus.Status400)
        res.Send("Invalid slug format")
        return
    }

    res.Send("Post slug: " + slug)
})

func isValidSlug(slug string) bool {
    // Simple slug validation: lowercase letters, numbers, hyphens
    for _, char := range slug {
        if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-') {
            return false
        }
    }
    return len(slug) > 0 && len(slug) <= 100
}
```

### Wildcard Parameters

```go
server.Get("/files/*", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // Wildcard captures the rest of the path
    if pathParam, exists := req.Segments.Get("*"); exists && len(pathParam) > 0 {
        filePath := pathParam[0]

        // Security: prevent path traversal
        if strings.Contains(filePath, "..") {
            res.Status(proteus.Status400)
            res.Send("Invalid file path")
            return
        }

        res.Send("File path: " + filePath)
    } else {
        res.Send("No file path provided")
    }
})

server.Get("/proxy/*", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    targetPath, _ := req.Segments.Get("*")

    // Forward the remaining path to another service
    res.Send("Proxying to: " + targetPath[0])
})

// Examples:
// /files/documents/report.pdf         -> filePath = "documents/report.pdf"
// /files/images/photos/vacation.jpg   -> filePath = "images/photos/vacation.jpg"
// /proxy/api/v1/users                 -> targetPath = "api/v1/users"
```

## Combining Query and Path Parameters

```go
server.Get("/users/:id/posts", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // Path parameter
    userID, _ := req.Segments.Get("id")

    // Query parameters for filtering and pagination
    page := 1
    if pageParam, exists := req.Params.Get("page"); exists && len(pageParam) > 0 {
        if p, err := strconv.Atoi(pageParam[0]); err == nil && p > 0 {
            page = p
        }
    }

    limit := 10
    if limitParam, exists := req.Params.Get("limit"); exists && len(limitParam) > 0 {
        if l, err := strconv.Atoi(limitParam[0]); err == nil && l > 0 && l <= 50 {
            limit = l
        }
    }

    status := "published"
    if statusParam, exists := req.Params.Get("status"); exists && len(statusParam) > 0 {
        validStatuses := []string{"draft", "published", "archived"}
        for _, valid := range validStatuses {
            if statusParam[0] == valid {
                status = statusParam[0]
                break
            }
        }
    }

    response := fmt.Sprintf("Posts by user %s: page %d, limit %d, status %s",
        userID[0], page, limit, status)
    res.Send(response)
})

// Example: /users/123/posts?page=2&limit=20&status=draft
```

## Parameter Processing Helpers

```go
// Helper functions for common parameter processing
func getIntParam(params *proteus.Params, name string, defaultValue int, min, max int) int {
    if param, exists := params.Get(name); exists && len(param) > 0 {
        if value, err := strconv.Atoi(param[0]); err == nil {
            if value >= min && value <= max {
                return value
            }
        }
    }
    return defaultValue
}

func getStringParam(params *proteus.Params, name, defaultValue string) string {
    if param, exists := params.Get(name); exists && len(param) > 0 {
        return param[0]
    }
    return defaultValue
}

func getBoolParam(params *proteus.Params, name string) bool {
    if param, exists := params.Get(name); exists {
        if len(param) == 0 {
            return true // Parameter exists without value
        }
        value := param[0]
        return value == "true" || value == "1" || value == "yes"
    }
    return false
}

func getFloatParam(params *proteus.Params, name string, defaultValue float64) float64 {
    if param, exists := params.Get(name); exists && len(param) > 0 {
        if value, err := strconv.ParseFloat(param[0], 64); err == nil {
            return value
        }
    }
    return defaultValue
}

// Usage in handler
server.Get("/search", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    query := getStringParam(req.Params, "q", "")
    page := getIntParam(req.Params, "page", 1, 1, 1000)
    limit := getIntParam(req.Params, "limit", 10, 1, 100)
    includeArchived := getBoolParam(req.Params, "include_archived")
    minScore := getFloatParam(req.Params, "min_score", 0.0)

    response := fmt.Sprintf("Search: %s, page: %d, limit: %d, archived: %t, min_score: %.2f",
        query, page, limit, includeArchived, minScore)
    res.Send(response)
})
```

## Advanced Parameter Handling

### Array Parameters

```go
server.Get("/filter", func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // Handle array parameters (same name multiple times)
    categories, hasCategories := req.Params.Get("category")
    tags, hasTags := req.Params.Get("tag")

    filters := map[string]interface{}{}

    if hasCategories {
        filters["categories"] = categories
    }

    if hasTags {
        filters["tags"] = tags
    }

    // Handle comma-separated values in single parameter
    if colors, exists := req.Params.Get("colors"); exists && len(colors) > 0 {
        colorList := strings.Split(colors[0], ",")
        filters["colors"] = colorList
    }

    res.AddHeader("Content-Type", "application/json")
    res.Send(fmt.Sprintf(`{"filters": %+v}`, filters))
})

// Examples:
// /filter?category=electronics&category=computers&tag=laptop&tag=gaming
// /filter?colors=red,blue,green
```

### Complex Parameter Validation

```go
// Parameter validation middleware
func validateSearchParams(req *proteus.HttpRequest, res *proteus.HttpResponse, next func()) {
    errors := []string{}

    // Validate search query
    if query, exists := req.Params.Get("q"); exists && len(query) > 0 {
        if len(query[0]) < 2 {
            errors = append(errors, "Search query must be at least 2 characters")
        }
        if len(query[0]) > 100 {
            errors = append(errors, "Search query must be less than 100 characters")
        }
    }

    // Validate page number
    if pageParam, exists := req.Params.Get("page"); exists && len(pageParam) > 0 {
        if page, err := strconv.Atoi(pageParam[0]); err != nil || page <= 0 {
            errors = append(errors, "Page must be a positive number")
        }
    }

    // Validate limit
    if limitParam, exists := req.Params.Get("limit"); exists && len(limitParam) > 0 {
        if limit, err := strconv.Atoi(limitParam[0]); err != nil || limit <= 0 || limit > 100 {
            errors = append(errors, "Limit must be between 1 and 100")
        }
    }

    // Validate sort field
    if sortParam, exists := req.Params.Get("sort"); exists && len(sortParam) > 0 {
        validSorts := []string{"name", "date", "price", "rating"}
        isValid := false
        for _, valid := range validSorts {
            if sortParam[0] == valid {
                isValid = true
                break
            }
        }
        if !isValid {
            errors = append(errors, "Invalid sort field")
        }
    }

    if len(errors) > 0 {
        res.Status(proteus.Status400)
        res.AddHeader("Content-Type", "application/json")
        res.Send(fmt.Sprintf(`{"errors": ["%s"]}`, strings.Join(errors, "\", \"")))
        return
    }

    next()
}

server.Get("/search", validateSearchParams, func(req *proteus.HttpRequest, res *proteus.HttpResponse) {
    // Parameters are validated at this point
    query := getStringParam(req.Params, "q", "")
    page := getIntParam(req.Params, "page", 1, 1, 1000)
    limit := getIntParam(req.Params, "limit", 10, 1, 100)
    sort := getStringParam(req.Params, "sort", "name")

    res.Send(fmt.Sprintf("Valid search: %s (page %d, limit %d, sort %s)",
        query, page, limit, sort))
})
```

## Method Reference

### Params Collection Methods
- `Get(name)` ([]string, bool) - Get parameter values and existence check
- `Has(name)` bool - Check if parameter exists
- `Set(name, values)` - Set parameter (mainly for testing)
- `Delete(name)` - Remove parameter (mainly for testing)

### Usage Patterns
- **Query Parameters**: `req.Params` - From URL query string (?key=value)
- **Path Parameters**: `req.Segments` - From route patterns (/:param)
- **Multiple Values**: Parameters can have multiple values (array handling)
- **Validation**: Always validate parameters before use
- **Defaults**: Provide sensible defaults for optional parameters
- **Type Conversion**: Convert strings to appropriate types with error handling
