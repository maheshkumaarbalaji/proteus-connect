package test

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/citadelofcode/proteus/internal"
)

func setResponseField(t *testing.T, response *internal.HttpResponse, field string, value any) {
	t.Helper()
	rv := reflect.ValueOf(response).Elem()
	fv := rv.FieldByName(field)
	if !fv.IsValid() {
		t.Fatalf(internal.TextColor.Red("Field %s does not exist on HttpResponse"), field)
	}
	reflect.NewAt(fv.Type(), unsafe.Pointer(fv.UnsafeAddr())).Elem().Set(reflect.ValueOf(value))
}

type failingWriter struct{}

func (f *failingWriter) Write(p []byte) (int, error) {
	return 0, errors.New("write failure")
}

// Test case to validate HttpResponse.Write for text bodies including header serialization.
func Test_HttpResponse_Write_TextBody(t *testing.T) {
	server := NewTestServer(t)
	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)

	response := NewTestResponse(t, "1.1", server, writer)
	response.Status(internal.Status200)
	response.AddHeader("Content-Type", "text/plain")
	response.BodyBytes = []byte("hello world")
	response.AddHeader("Content-Length", "11")

	err := response.Write()
	if err != nil {
		t.Fatalf(internal.TextColor.Red("Expected Write to succeed, but received error: %v"), err)
	}

	writer.Flush()
	output := buffer.String()
	if !strings.Contains(output, "HTTP/1.1 200 OK") {
		t.Errorf(internal.TextColor.Red("Expected status line in response, got %q"), output)
	}
	if !strings.Contains(output, "hello world") {
		t.Errorf(internal.TextColor.Red("Expected body content in response, got %q"), output)
	}
}

// Test case to ensure Write returns an error when status code is missing.
func Test_HttpResponse_Write_StatusError(t *testing.T) {
	server := NewTestServer(t)
	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)

	response := NewTestResponse(t, "1.1", server, writer)
	response.BodyBytes = []byte("content without status")

	err := response.Write()
	if err == nil {
		t.Fatal(internal.TextColor.Red("Expected ResponseError when status code is missing"))
	}
	if _, ok := err.(*internal.ResponseError); !ok {
		t.Fatalf(internal.TextColor.Red("Expected ResponseError type, got %T"), err)
	}
}

// Test case to ensure empty response version triggers error.
func Test_HttpResponse_Write_EmptyVersion(t *testing.T) {
	server := NewTestServer(t)
	response := NewTestResponse(t, "1.1", server, bufio.NewWriter(&bytes.Buffer{}))
	setResponseField(t, response, "Version", "")

	response.Status(internal.Status200)
	err := response.Write()
	if err == nil {
		t.Fatal(internal.TextColor.Red("Expected error when response version is empty"))
	}
	if resErr, ok := err.(*internal.ResponseError); !ok || resErr.Section != "StatusLine" {
		t.Fatalf(internal.TextColor.Red("Expected StatusLine error, got %T"), err)
	}
}

// Test case to cover binary body writing path when Content-Type header is absent.
func Test_HttpResponse_Write_BinaryWithoutContentType(t *testing.T) {
	server := NewTestServer(t)
	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)

	response := NewTestResponse(t, "1.1", server, writer)
	response.Status(internal.Status200)
	response.BodyBytes = []byte{0xde, 0xad, 0xbe, 0xef}
	response.AddHeader("Content-Length", "4")

	err := response.Write()
	if err != nil {
		t.Fatalf(internal.TextColor.Red("Expected Write to succeed for binary body, received error: %v"), err)
	}
	writer.Flush()

	output := buffer.Bytes()
	if !bytes.Contains(output, []byte{0xde, 0xad, 0xbe, 0xef}) {
		t.Errorf(internal.TextColor.Red("Expected binary payload in response, got %v"), output)
	}
}

// Test case to ensure SetServer assigns the provided server reference.
func Test_HttpResponse_SetServer(t *testing.T) {
	server := NewTestServer(t)
	response := NewTestResponse(t, "1.1", server, nil)

	if response.Server != server {
		t.Error(internal.TextColor.Red("Expected response to reference server from helper"))
	}

	another := NewTestServer(t)
	response.SetServer(another)
	if response.Server != another {
		t.Error(internal.TextColor.Red("Expected SetServer to update server reference"))
	}
}

// Test case to cover Write error when writer is missing.
func Test_HttpResponse_Write_NoWriter(t *testing.T) {
	server := NewTestServer(t)
	response := NewTestResponse(t, "1.1", server, bufio.NewWriter(&bytes.Buffer{}))
	setResponseField(t, response, "writer", (*bufio.Writer)(nil))
	response.Status(internal.Status200)

	err := response.Write()
	if err == nil {
		t.Fatal(internal.TextColor.Red("Expected ResponseError when writer is not initialized"))
	}
}

// Test case to cover status line write errors bubbling up correctly.
func Test_HttpResponse_Write_StatusLineWriteError(t *testing.T) {
	server := NewTestServer(t)
	response := NewTestResponse(t, "1.1", server, bufio.NewWriter(&bytes.Buffer{}))
	setResponseField(t, response, "writer", bufio.NewWriterSize(&failingWriter{}, 1))
	response.Status(internal.Status200)

	err := response.Write()
	if err == nil {
		t.Fatal(internal.TextColor.Red("Expected ResponseError when writer fails during status line write"))
	}
	if resErr, ok := err.(*internal.ResponseError); ok {
		if resErr.Section != "StatusLine" {
			t.Errorf(internal.TextColor.Red("Expected StatusLine error section, got %s"), resErr.Section)
		}
	} else {
		t.Fatalf(internal.TextColor.Red("Expected ResponseError, got %T"), err)
	}
}

// Test case to validate HTTP/0.9 write path which skips status line and headers.
func Test_HttpResponse_Write_HTTP09(t *testing.T) {
	server := NewTestServer(t)
	var buffer bytes.Buffer
	response := NewTestResponse(t, "0.9", server, bufio.NewWriter(&buffer))
	response.BodyBytes = []byte("legacy body")

	err := response.Write()
	if err != nil {
		t.Fatalf(internal.TextColor.Red("Expected HTTP/0.9 write to succeed, got %v"), err)
	}

	if buffer.String() != "legacy body" {
		t.Errorf(internal.TextColor.Red("Expected body only output, got %q"), buffer.String())
	}
}

// Test case to validate SendFile metadata-only path skips loading body.
func Test_HttpResponse_SendFile_MetadataOnly(t *testing.T) {
	server := NewTestServer(t)
	root := t.TempDir()
	filePath := filepath.Join(root, "data.txt")
	if err := os.WriteFile(filePath, []byte("meta"), 0644); err != nil {
		t.Fatalf(internal.TextColor.Red("Error creating temp file: %v"), err)
	}

	var buffer bytes.Buffer
	response := NewTestResponse(t, "1.1", server, bufio.NewWriter(&buffer))
	response.Status(internal.Status200)
	if err := response.SendFile(filePath, true); err != nil {
		t.Fatalf(internal.TextColor.Red("Expected SendFile metadata path to succeed, got %v"), err)
	}

	if len(response.BodyBytes) != 0 {
		t.Error(internal.TextColor.Red("Expected body bytes to remain empty in metadata-only response"))
	}
}
