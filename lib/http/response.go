package http

import (
	"bufio"
	"errors"
	"fmt"
	"net/textproto"
	"slices"
	"strconv"
	"strings"
	"time"
	"github.com/maheshkumaarbalaji/proteus/lib/fs"
)

// Structure to represent a HTTP response sent back by the server to the client.
type HttpResponse struct {
	StatusCode int
	StatusMessage string
	Version string
	Headers Headers
	Body []byte
	writer *bufio.Writer
}

func (res *HttpResponse) initialize(version string) {
	res.Version = strings.TrimSpace(version)
	res.Headers = make(Headers)
	res.addGeneralHeaders()
	res.addResponseHeaders()
	res.Version = GetHighestVersion()
}

func (res *HttpResponse) setWriter(writer *bufio.Writer) {
	res.writer = writer
}

func (res *HttpResponse) addGeneralHeaders() {
	res.Headers.Add("Date", getRfc1123Time())
}

func (res *HttpResponse) addResponseHeaders() {
	res.Headers.Add("Server", GetServerName())
}

func (res *HttpResponse) write() {
	err := res.writeStatusLine()
	if err != nil {
		LogError(err.Error())
		return
	}

	err = res.writeHeaders()
	if err != nil {
		LogError(err.Error())
		return
	}

	err = res.writeBody()
	if err != nil {
		LogError(err.Error())
		return
	}

	err = res.writer.Flush()
	if err != nil {
		LogError(err.Error())
		return
	}
}

func (res *HttpResponse) writeStatusLine() error {
	if res.writer == nil {
		return errors.New("error occurred while writing response status line: writer object not initialized")
	}

	if res.StatusCode == 0 {
		return errors.New("error occurred while writing response status line: status code cannot be zero")
	}

	if res.Version == "" {
		return errors.New("error occurred while writing response status line: Protocol version not set")
	}

	_, err := res.writer.WriteString(fmt.Sprintf("HTTP/%s %d %s%s", res.Version, res.StatusCode, res.StatusMessage, HEADER_LINE_SEPERATOR))
	if err != nil {
		return errors.New("error occurred while writing response status line: " + err.Error())
	}

	return nil
}

func (res *HttpResponse) writeHeaders() error {
	if res.writer == nil {
		return errors.New("error occurred while writing response headers: writer object not initialized")
	}

	for key, values := range res.Headers {
		value := strings.Join(values, ",")
		_, err := res.writer.WriteString(fmt.Sprintf("%s: %s%s", key, value, HEADER_LINE_SEPERATOR))
		if err != nil {
			return errors.New("error occurred while writing response headers: " + err.Error())
		}
	}
	res.writer.WriteString(HEADER_LINE_SEPERATOR)

	return nil
}

func (res *HttpResponse) writeBody() error {
	if res.writer == nil {
		return errors.New("error occurred while writing response body: writer object not initialized")
	}

	if len(res.Body) > 0 {
		ContentType, exists := res.Headers.Get("Content-Type")
		if exists {
			ContentType = strings.TrimSpace(ContentType)
			ContentType = strings.ToLower(ContentType)
			if strings.HasPrefix(ContentType, "text") {
				_, err := res.writer.WriteString(string(res.Body))
				if err != nil {
					return errors.New("error occurred while writing response body: " + err.Error())
				}
			} else {
				_, err := res.writer.Write(res.Body)
				if err != nil {
					return errors.New("error occurred while writing response body: " + err.Error())
				}
			}
		}
	}

	return nil
}

// Adds a new key-value pair to the request headers collection.
func (res *HttpResponse) AddHeader(HeaderKey string, HeaderValue string) {
	if slices.Contains(DateHeaders, textproto.CanonicalMIMEHeaderKey(HeaderKey)) {
		_, err := time.Parse(time.RFC1123, HeaderValue)
		_, errOne := time.Parse(time.ANSIC, HeaderValue)

		if err == nil || errOne == nil {
			res.Headers.Add(HeaderKey, HeaderValue)
		}
	} else {
		res.Headers.Add(HeaderKey, HeaderValue)
	}
}

// Sets the status of the HTTP response instance.
func (res *HttpResponse) Status(status StatusCode) {
	res.StatusCode = int(status)
	res.StatusMessage = status.GetStatusMessage()
}

// Send the given file from the local file system as the HTTP response.
func (res *HttpResponse) SendFile(CompleteFilePath string, OnlyMetadata bool) {
	fileMediaType, exists := GetContentType(CompleteFilePath)
	if exists {
		file, err := fs.GetFile(CompleteFilePath, fileMediaType, OnlyMetadata)
		if err == nil {
			res.AddHeader("Content-Type", fileMediaType)
			res.AddHeader("Content-Length", strconv.FormatInt(file.Size, 10))
			res.AddHeader("Last-Modified", file.LastModifiedAt.Format(time.RFC1123))
			if !OnlyMetadata {
				res.Body = file.Contents
			}
			
			res.write()
		}
	}
}

// Sends a the given error content as response back to the client.
func (res *HttpResponse) SendError(Content string) {
	responseContent := []byte(Content)
	res.AddHeader("Content-Type", ERROR_MSG_CONTENT_TYPE)
	res.AddHeader("Content-Length", strconv.Itoa(len(responseContent)))
	res.Body = responseContent
	res.write()
}