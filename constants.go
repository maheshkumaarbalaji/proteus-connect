// Package proteus provides a versatile web server framework for Go, inspired by Node.js Express.
// It offers a simple and intuitive API for building HTTP servers with support for routing,
// middleware, request/response handling, and various HTTP status codes and logging formats.
package proteus

import (
	"github.com/citadelofcode/proteus/internal"
)

// HTTP status codes supported by the proteus web server.
// These constants provide convenient access to all standard HTTP status codes
// as defined by IANA HTTP Status Code Registry.
const (
	// 1xx Informational responses - indicate that the request was received and understood
	Status100 StatusCode = internal.Status100 // Continue
	Status101 StatusCode = internal.Status101 // Switching Protocols
	Status102 StatusCode = internal.Status102 // Processing
	Status103 StatusCode = internal.Status103 // Early Hints

	// 2xx Success - indicate that the action requested by the client was received, understood, and accepted
	Status200 StatusCode = internal.Status200 // OK
	Status201 StatusCode = internal.Status201 // Created
	Status202 StatusCode = internal.Status202 // Accepted
	Status203 StatusCode = internal.Status203 // Non-Authoritative Information
	Status204 StatusCode = internal.Status204 // No Content
	Status205 StatusCode = internal.Status205 // Reset Content
	Status206 StatusCode = internal.Status206 // Partial Content
	Status207 StatusCode = internal.Status207 // Multi-Status
	Status208 StatusCode = internal.Status208 // Already Reported

	// 3xx Redirection - indicate that further action needs to be taken by the client to complete the request
	Status300 StatusCode = internal.Status300 // Multiple Choices
	Status301 StatusCode = internal.Status301 // Moved Permanently
	Status302 StatusCode = internal.Status302 // Found
	Status303 StatusCode = internal.Status303 // See Other
	Status304 StatusCode = internal.Status304 // Not Modified
	Status305 StatusCode = internal.Status305 // Use Proxy
	Status307 StatusCode = internal.Status307 // Temporary Redirect
	Status308 StatusCode = internal.Status308 // Permanent Redirect

	// 4xx Client Error - indicate that the client seems to have made an error
	Status400 StatusCode = internal.Status400 // Bad Request
	Status401 StatusCode = internal.Status401 // Unauthorized
	Status402 StatusCode = internal.Status402 // Payment Required
	Status403 StatusCode = internal.Status403 // Forbidden
	Status404 StatusCode = internal.Status404 // Not Found
	Status405 StatusCode = internal.Status405 // Method Not Allowed
	Status406 StatusCode = internal.Status406 // Not Acceptable
	Status407 StatusCode = internal.Status407 // Proxy Authentication Required
	Status408 StatusCode = internal.Status408 // Request Timeout
	Status409 StatusCode = internal.Status409 // Conflict
	Status410 StatusCode = internal.Status410 // Gone
	Status411 StatusCode = internal.Status411 // Length Required
	Status412 StatusCode = internal.Status412 // Precondition Failed
	Status413 StatusCode = internal.Status413 // Payload Too Large
	Status414 StatusCode = internal.Status414 // URI Too Long
	Status415 StatusCode = internal.Status415 // Unsupported Media Type
	Status416 StatusCode = internal.Status416 // Range Not Satisfiable
	Status417 StatusCode = internal.Status417 // Expectation Failed
	Status421 StatusCode = internal.Status421 // Misdirected Request
	Status422 StatusCode = internal.Status422 // Unprocessable Entity
	Status423 StatusCode = internal.Status423 // Locked
	Status424 StatusCode = internal.Status424 // Failed Dependency
	Status425 StatusCode = internal.Status425 // Too Early
	Status426 StatusCode = internal.Status426 // Upgrade Required
	Status428 StatusCode = internal.Status428 // Precondition Required
	Status429 StatusCode = internal.Status429 // Too Many Requests
	Status431 StatusCode = internal.Status431 // Request Header Fields Too Large

	// 5xx Server Error - indicate that the server failed to fulfill a valid request
	Status500 StatusCode = internal.Status500 // Internal Server Error
	Status501 StatusCode = internal.Status501 // Not Implemented
	Status502 StatusCode = internal.Status502 // Bad Gateway
	Status503 StatusCode = internal.Status503 // Service Unavailable
	Status504 StatusCode = internal.Status504 // Gateway Timeout
	Status505 StatusCode = internal.Status505 // HTTP Version Not Supported
	Status506 StatusCode = internal.Status506 // Variant Also Negotiates
	Status507 StatusCode = internal.Status507 // Insufficient Storage
	Status508 StatusCode = internal.Status508 // Loop Detected
	Status511 StatusCode = internal.Status511 // Network Authentication Required
)

// Logging levels available for server logs.
// These constants define the severity levels for log messages output by the server.
const (
	// INFO_LEVEL represents informational messages that highlight normal server operation.
	// Use this level for general operational messages that might be useful for debugging.
	INFO_LEVEL = internal.INFO_LEVEL

	// ERROR_LEVEL represents error conditions that require immediate attention.
	// Use this level for exceptions, failures, and other critical issues.
	ERROR_LEVEL = internal.ERROR_LEVEL

	// WARN_LEVEL represents potentially harmful situations that don't stop the server.
	// Use this level for deprecated API usage, recoverable errors, or unusual conditions.
	WARN_LEVEL = internal.WARN_LEVEL
)

// Logging formats for HTTP request processing logs.
// These constants define different predefined formats for logging HTTP requests and responses,
// similar to Apache and other web server logging formats.
const (
	// COMMON_LOGGER uses the Common Log Format (CLF) as defined in Apache HTTP Server.
	// Format: :remote-addr [:date[clf]] ":method :url HTTP/:http-version" :status :res[content-length]
	// Example: 127.0.0.1 [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326
	COMMON_LOGGER = internal.COMMON_LOGGER

	// DEV_LOGGER provides a concise format suitable for development environments.
	// Format: :method :url :status :response-time ms - :res[content-length]
	// Example: GET / 200 15.456 ms - 1234
	DEV_LOGGER = internal.DEV_LOGGER

	// TINY_LOGGER provides the most minimal logging format.
	// Format: :method :url :status :res[content-length] - :response-time ms
	// Example: GET / 200 1234 - 15.456 ms
	TINY_LOGGER = internal.TINY_LOGGER

	// SHORT_LOGGER provides a balance between information and brevity.
	// Format: :remote-addr :method :url HTTP/:http-version :status :res[content-length] - :response-time ms
	// Example: 127.0.0.1 GET / HTTP/1.1 200 1234 - 15.456 ms
	SHORT_LOGGER = internal.SHORT_LOGGER
)

// TextColor provides utility functions to apply ANSI color codes to text output.
// This is useful for enhancing the readability of console logs with colored text
// on terminals that support ANSI escape sequences.
var TextColor = internal.TextColor
