package test

import (
	"bufio"
	"bytes"
	"errors"
	"reflect"
	"testing"
	"unsafe"

	"github.com/citadelofcode/proteus/internal"
)

//go:linkname httpResponseWriteHeaders github.com/citadelofcode/proteus/internal.(*HttpResponse).writeHeaders
func httpResponseWriteHeaders(*internal.HttpResponse) error

//go:linkname httpResponseWriteBody github.com/citadelofcode/proteus/internal.(*HttpResponse).writeBody
func httpResponseWriteBody(*internal.HttpResponse) error

type errWriter struct{}

func (errWriter) Write(p []byte) (int, error) {
	return 0, errors.New("forced failure")
}

func setResponseWriter(response *internal.HttpResponse, writer *bufio.Writer) {
	rv := reflect.ValueOf(response).Elem()
	fv := rv.FieldByName("writer")
	reflect.NewAt(fv.Type(), unsafe.Pointer(fv.UnsafeAddr())).Elem().Set(reflect.ValueOf(writer))
}

// Test case to validate header write errors surface as ResponseError.
func Test_HttpResponse_writeHeaders_Error(t *testing.T) {
	server := NewTestServer(t)
	response := NewTestResponse(t, "1.1", server, bufio.NewWriter(&bytes.Buffer{}))
	response.Headers.Add("X-Test", "value")
	setResponseWriter(response, bufio.NewWriterSize(errWriter{}, 1))

	err := httpResponseWriteHeaders(response)
	if err == nil {
		t.Fatal(internal.TextColor.Red("Expected error when underlying writer fails"))
	}
	if resErr, ok := err.(*internal.ResponseError); !ok || resErr.Section != "Header" {
		t.Fatalf(internal.TextColor.Red("Expected header ResponseError, got %T (%v)"), err, err)
	}
}

// Test case to validate writeHeaders successfully writes terminating CRLF when no headers exist.
func Test_HttpResponse_writeHeaders_Empty(t *testing.T) {
	server := NewTestServer(t)
	response := NewTestResponse(t, "1.1", server, bufio.NewWriter(&bytes.Buffer{}))
	response.Headers = make(internal.Headers)
	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)
	setResponseWriter(response, writer)

	if err := httpResponseWriteHeaders(response); err != nil {
		t.Fatalf(internal.TextColor.Red("Expected writeHeaders to succeed, got %v"), err)
	}
	writer.Flush()
	if buffer.String() != "\r\n" {
		t.Errorf(internal.TextColor.Red("Expected terminating CRLF, got %q"), buffer.String())
	}
}

// Test case to validate body write errors for text responses.
func Test_HttpResponse_writeBody_Error(t *testing.T) {
	server := NewTestServer(t)
	response := NewTestResponse(t, "1.1", server, bufio.NewWriter(&bytes.Buffer{}))
	response.BodyBytes = []byte("content")
	response.AddHeader("Content-Type", "text/plain")
	setResponseWriter(response, bufio.NewWriterSize(errWriter{}, 1))

	err := httpResponseWriteBody(response)
	if err == nil {
		t.Fatal(internal.TextColor.Red("Expected error when writing body fails"))
	}
	if resErr, ok := err.(*internal.ResponseError); !ok || resErr.Section != "Body" {
		t.Fatalf(internal.TextColor.Red("Expected body ResponseError, got %T (%v)"), err, err)
	}
}

// Test case to ensure writeBody succeeds when there is no body content.
func Test_HttpResponse_writeBody_NoContent(t *testing.T) {
	server := NewTestServer(t)
	response := NewTestResponse(t, "1.1", server, bufio.NewWriter(&bytes.Buffer{}))
	response.BodyBytes = nil

	if err := httpResponseWriteBody(response); err != nil {
		t.Fatalf(internal.TextColor.Red("Expected nil error when body is empty, got %v"), err)
	}
}

// Test case to validate Content-Type normalization with surrounding whitespace and mixed case.
func Test_HttpResponse_writeBody_TextWhitespace(t *testing.T) {
	server := NewTestServer(t)
	buffer := &bytes.Buffer{}
	response := NewTestResponse(t, "1.1", server, buffer)
	response.BodyBytes = []byte(" spaced ")
	response.AddHeader("Content-Type", "  TEXT/PLAIN  ")

	if err := httpResponseWriteBody(response); err != nil {
		t.Fatalf(internal.TextColor.Red("Expected body write to succeed with normalized content type, got %v"), err)
	}
}

// Test case to validate binary content path when Content-Type is present.
func Test_HttpResponse_writeBody_BinaryContentType(t *testing.T) {
	server := NewTestServer(t)
	response := NewTestResponse(t, "1.1", server, bufio.NewWriter(&bytes.Buffer{}))
	response.BodyBytes = []byte{0x1, 0x2, 0x3}
	response.AddHeader("Content-Type", "application/octet-stream")

	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)
	setResponseWriter(response, writer)

	if err := httpResponseWriteBody(response); err != nil {
		t.Fatalf(internal.TextColor.Red("Expected binary write to succeed, got %v"), err)
	}
	writer.Flush()
	if !bytes.Contains(buffer.Bytes(), []byte{0x1, 0x2, 0x3}) {
		t.Errorf(internal.TextColor.Red("Expected binary payload to be written, got %v"), buffer.Bytes())
	}
}
