package internal

import (
	"bufio"
	"fmt"
	"io"
	"net/textproto"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Structure to represent a HTTP response sent back by the server to the client.
type HttpResponse struct {
	// Status code of the response being sent back to the client like 200, 203, 404 etc.
	StatusCode int
	// Status message associated with the response status code.
	StatusMessage string
	// HTTP version of the response being sent back.
	Version string
	// Collection of all response headers being sent by the server.
	Headers Headers
	// Complete contents of the response body as a stream of bytes.
	BodyBytes []byte
	// Streamed writer instance to write the response bytes to the network stream.
	writer *bufio.Writer
	// The server instance processing this response.
	Server *HttpServer
	// key-value pairs to hold variables available during the entire response lifecycle.
	Locals map[string]any
	// FileSystem instance to access the local file system.
	fs *FileSystem
}

// Initialize sets up an HttpResponse instance with default values and prepares it for HTTP output.
// Use for creating new response instances with proper HTTP version handling.
// Sets HTTP version (defaults to "0.9"), creates header/locals collections, adds standard headers.
// Parameters: version (HTTP version), writer (output destination).
// Returns: Configured response ready for sending data.
func (res *HttpResponse) Initialize(version string, writer io.Writer) {
	version = strings.TrimSpace(version)
	if version == "" {
		res.Version = "0.9"
	} else {
		res.Version = version
	}
	res.Headers = make(Headers)
	res.Locals = make(map[string]any)
	res.addGeneralHeaders()
	res.addResponseHeaders()
	res.writer = bufio.NewWriter(writer)
	res.Server = nil
	res.fs = new(FileSystem)
}

// Sets the server field to the given server instance reference.
func (res *HttpResponse) SetServer(serverRef *HttpServer) {
	res.Server = serverRef
}

// Adds all the general HTTP headers to the HttpResponse instance.
// Headers are added only if the given HttpResponse object is not a test instance and the response version is not HTTP/0.9.
func (res *HttpResponse) addGeneralHeaders() {
	if !strings.EqualFold(res.Version, "0.9") {
		res.Headers.Add("Date", GetRfc1123Time())
	}
}

// Adds all the default response HTTP headers to the HttpResponse instance.
// Headers are added only if the given HttpResponse object is not a test instance and the response version is not HTTP/0.9.
func (res *HttpResponse) addResponseHeaders() {
	if !strings.EqualFold(res.Version, "0.9") {
		res.Headers.Add("Server", GetServerDefaults("server_name").(string))
	}
}

// Write transmits the complete HTTP response (status, headers, body) to the client connection.
// Use to send finalized response data to network writer with proper HTTP formatting.
// Handles HTTP/0.9 (body only) and HTTP/1.0+ (full response) versions automatically.
// Called internally by Send(), SendFile(), and SendError() methods.
// Returns: error if transmission fails, nil on successful write and flush.
func (res *HttpResponse) Write() error {
	if res.writer == nil {
		resErr := new(ResponseError)
		resErr.Section = "RespWrite"
		resErr.Value = ""
		resErr.Message = "Writer object not initialized"
		return resErr
	}

	var err error
	if !strings.EqualFold(res.Version, "0.9") {
		err = res.writeStatusLine()
		if err != nil {
			return err
		}
	}

	if !strings.EqualFold(res.Version, "0.9") {
		err = res.writeHeaders()
		if err != nil {
			return err
		}
	}

	err = res.writeBody()
	if err != nil {
		return err
	}

	err = res.writer.Flush()
	if err != nil {
		resErr := new(ResponseError)
		resErr.Section = "RespWrite"
		resErr.Value = ""
		resErr.Message = "Writer object could not be flushed"
		return resErr
	}

	return nil
}

// Writes the HTTP response status line to the response byte stream.
func (res *HttpResponse) writeStatusLine() error {
	if res.StatusCode == 0 {
		resErr := new(ResponseError)
		resErr.Section = "StatusLine"
		resErr.Value = ""
		resErr.Message = "Status code for the response cannot be zero"
		return resErr
	}

	if res.Version == "" {
		resErr := new(ResponseError)
		resErr.Section = "StatusLine"
		resErr.Value = ""
		resErr.Message = "Response version cannot be empty"
		return resErr
	}

	_, err := res.writer.WriteString(fmt.Sprintf("HTTP/%s %d %s%s", res.Version, res.StatusCode, res.StatusMessage, HEADER_LINE_SEPERATOR))
	if err != nil {
		resErr := new(ResponseError)
		resErr.Section = "StatusLine"
		resErr.Value = ""
		resErr.Message = fmt.Sprintf("Error while writing response status line :: %s", err.Error())
		return resErr
	}

	return nil
}

// Writes the HTTP response headers to the response byte stream.
func (res *HttpResponse) writeHeaders() error {
	for key, values := range res.Headers {
		value := strings.Join(values, ",")
		_, err := res.writer.WriteString(fmt.Sprintf("%s: %s%s", key, value, HEADER_LINE_SEPERATOR))
		if err != nil {
			resErr := new(ResponseError)
			resErr.Section = "Header"
			resErr.Value = fmt.Sprintf("%s: %s", key, value)
			resErr.Message = fmt.Sprintf("Error while writing response header :: %s", err.Error())
			return resErr
		}
	}

	_, err := res.writer.WriteString(HEADER_LINE_SEPERATOR)
	if err != nil {
		resErr := new(ResponseError)
		resErr.Section = "Header"
		resErr.Value = HEADER_LINE_SEPERATOR
		resErr.Message = fmt.Sprintf("Error while writing response header :: %s", err.Error())
		return resErr
	}

	return nil
}

// Writes the response body to the response byte stream.
func (res *HttpResponse) writeBody() error {
	if len(res.BodyBytes) > 0 {
		ContentType, exists := res.Headers.Get("Content-Type")
		if exists {
			ContentType = strings.TrimSpace(ContentType)
			ContentType = strings.ToLower(ContentType)
			if strings.HasPrefix(ContentType, "text") {
				_, err := res.writer.WriteString(string(res.BodyBytes))
				if err != nil {
					resErr := new(ResponseError)
					resErr.Section = "Body"
					resErr.Value = ContentType
					resErr.Message = fmt.Sprintf("Error while writing response body :: %s", err.Error())
					return resErr
				}
			} else {
				_, err := res.writer.Write(res.BodyBytes)
				if err != nil {
					resErr := new(ResponseError)
					resErr.Section = "Body"
					resErr.Value = ContentType
					resErr.Message = fmt.Sprintf("Error while writing response body :: %s", err.Error())
					return resErr
				}
			}
		} else {
			_, err := res.writer.Write(res.BodyBytes)
			if err != nil {
				resErr := new(ResponseError)
				resErr.Section = "Body"
				resErr.Value = "Body Write without Content Type"
				resErr.Message = fmt.Sprintf("Error while writing response body :: %s", err.Error())
				return resErr
			}
		}
	}

	return nil
}

// AddHeader adds a header key-value pair to response headers with validation and normalization.
// Use to set HTTP headers like Content-Type, Cache-Control, or custom headers before sending.
// Normalizes keys to canonical format, validates date headers, supports multiple values per key.
// Parameters: HeaderKey (header name), HeaderValue (header value).
// Returns: No return value, headers stored in response Headers collection.
//
// Common response headers:
//   - Content-Type, Content-Length, Content-Encoding
//   - Cache-Control, Expires, ETag, Last-Modified
//   - Location (for redirects), Set-Cookie
//   - Access-Control-* (for CORS)
func (res *HttpResponse) AddHeader(HeaderKey string, HeaderValue string) {
	if slices.Contains(DateHeaders, textproto.CanonicalMIMEHeaderKey(HeaderKey)) {
		isValid, _ := IsHttpDate(HeaderValue)
		if isValid {
			res.Headers.Add(HeaderKey, HeaderValue)
		} else {
			res.Server.Log(fmt.Sprintf("Error while adding header - [%s] :: Date string must conform to one of these formats - RFC1123 or ANSIC", HeaderKey), ERROR_LEVEL)
		}
	} else {
		res.Headers.Add(HeaderKey, HeaderValue)
	}
}

// Status sets the HTTP status code and message for this response.
// Use to set response status before sending content (200 OK, 404 Not Found, 500 Error, etc.).
// Configures both numeric code and text message for proper HTTP response line.
// Parameter: status (StatusCode constant like Status200, Status404, Status500).
// Returns: No return value, status stored in response for transmission.
func (res *HttpResponse) Status(status StatusCode) {
	res.StatusCode = int(status)
	res.StatusMessage = status.GetStatusMessage()
}

// SendFile transmits a file from filesystem as HTTP response with automatic headers and MIME detection.
// Use to serve static files, downloads, or handle HEAD requests for file metadata.
// Sets Content-Length, Last-Modified, and Content-Type headers automatically.
// Parameters: CompleteFilePath (absolute path), OnlyMetadata (true for HEAD requests).
// Returns: error on file access/transmission failure, nil on success.
func (res *HttpResponse) SendFile(CompleteFilePath string, OnlyMetadata bool) error {
	file, err := res.fs.GetFile(CompleteFilePath)
	if err != nil {
		return err
	}

	res.Headers.Add("Content-Length", strconv.FormatInt(file.Size(), 10))
	res.Headers.Add("Last-Modified", file.LastModified().Format(time.RFC1123))
	_, ok := res.Headers.Get("Content-Type")
	if !ok {
		res.Headers.Add("Content-Type", file.MediaType())
	}

	if !OnlyMetadata {
		contents, err := file.Contents()
		if err != nil {
			return err
		}
		res.BodyBytes = contents
	}

	return res.Write()
}

// SendError transmits error content as HTTP response with proper error headers and content type.
// Use to send error messages, HTML error pages, or JSON error responses to clients.
// Sets appropriate Content-Type and Content-Length headers automatically.
// Parameter: Content (error message or formatted content to send).
// Returns: error on transmission failure, nil on successful error response.
func (res *HttpResponse) SendError(Content string) error {
	responseContent := []byte(Content)
	res.Headers.Add("Content-Type", ERROR_MSG_CONTENT_TYPE)
	res.Headers.Add("Content-Length", strconv.Itoa(len(responseContent)))
	res.BodyBytes = responseContent
	return res.Write()
}

// Send transmits string content as HTTP response body with automatic headers.
// Use to send text, JSON, HTML, or other string-based content to clients.
// Sets Content-Length automatically and uses default Content-Type if not specified.
// Parameter: content (string data to send as response body).
// Returns: error on transmission failure, nil on successful response.
func (res *HttpResponse) Send(content string) error {
	_, ok := res.Headers.Get("Content-Type")
	if !ok {
		fileMediaType := GetServerDefaults("content_type").(string)
		res.Headers.Add("Content-Type", fileMediaType)
	}

	content = strings.TrimSpace(content)
	contentBuffer := []byte(content)
	res.Headers.Add("Content-Length", strconv.Itoa(len(contentBuffer)))
	res.BodyBytes = contentBuffer
	return res.Write()
}
