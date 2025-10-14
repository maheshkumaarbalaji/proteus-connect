package internal

import (
	"bytes"
	"html/template"
)

type StatusCode int

// Structure to represent a response status code and its associated information.
type HttpStatus struct {
	// HTTP response status code.
	Code StatusCode
	// Short message for the corresponding status code.
	Message string
	// Error description for error status codes (>=400).
	ErrorDescription string
}

const (
	Status100 StatusCode = 100
	Status101 StatusCode = 101
	Status102 StatusCode = 102
	Status103 StatusCode = 103

	Status200 StatusCode = 200
	Status201 StatusCode = 201
	Status202 StatusCode = 202
	Status203 StatusCode = 203
	Status204 StatusCode = 204
	Status205 StatusCode = 205
	Status206 StatusCode = 206
	Status207 StatusCode = 300
	Status208 StatusCode = 208

	Status300 StatusCode = 300
	Status301 StatusCode = 301
	Status302 StatusCode = 302
	Status303 StatusCode = 303
	Status304 StatusCode = 304
	Status305 StatusCode = 305
	Status307 StatusCode = 307
	Status308 StatusCode = 308

	Status400 StatusCode = 400
	Status401 StatusCode = 401
	Status402 StatusCode = 402
	Status403 StatusCode = 403
	Status404 StatusCode = 404
	Status405 StatusCode = 405
	Status406 StatusCode = 406
	Status407 StatusCode = 407
	Status408 StatusCode = 408
	Status409 StatusCode = 409
	Status410 StatusCode = 410
	Status411 StatusCode = 411
	Status412 StatusCode = 412
	Status413 StatusCode = 413
	Status414 StatusCode = 414
	Status415 StatusCode = 415
	Status416 StatusCode = 416
	Status417 StatusCode = 417
	Status421 StatusCode = 421
	Status422 StatusCode = 422
	Status423 StatusCode = 423
	Status424 StatusCode = 424
	Status425 StatusCode = 425
	Status426 StatusCode = 426
	Status428 StatusCode = 428
	Status429 StatusCode = 429
	Status431 StatusCode = 431

	Status500 StatusCode = 500
	Status501 StatusCode = 501
	Status502 StatusCode = 502
	Status503 StatusCode = 503
	Status504 StatusCode = 504
	Status505 StatusCode = 505
	Status506 StatusCode = 506
	Status507 StatusCode = 507
	Status508 StatusCode = 508
	Status511 StatusCode = 511
)

// GetStatusMessage retrieves the standard HTTP reason phrase for this status code.
// Used internally by the response system to construct HTTP status lines like "HTTP/1.1 200 OK".
// Returns the official reason phrase (e.g., "OK", "Not Found") for standard status codes.
// Returns empty string for unrecognized status codes.
func (code StatusCode) GetStatusMessage() string {
	for _, stat := range ResponseStatusCodes {
		if stat.Code == code {
			return stat.Message
		}
	}

	return ""
}

// GetErrorContent generates a default HTML error page for HTTP error status codes.
// Used when custom error handling is not implemented to provide standardized error pages.
// Creates complete HTML document with status code, reason phrase, and error description.
// Returns empty string if status code is not found in the registry.
func (code StatusCode) GetErrorContent() string {
	htmlTemplate := `<html>
					<head>
					<title>{{printf "%s - Response" .Code}</title>
					</head>
					<body>
					<h1>{{printf "%d - %s" .Code .Message}}</h1>
					<p>{{.ErrorDescription}}</p>
					</body>
				</html>`

	for _, stat := range ResponseStatusCodes {
		if code == stat.Code {
			temp, err := template.New("errorResponse").Parse(htmlTemplate)
			if err != nil {
				break
			}

			var tmpBytes bytes.Buffer
			err = temp.Execute(&tmpBytes, stat)
			if err != nil {
				break
			}

			return tmpBytes.String()
		}
	}

	return ""
}
