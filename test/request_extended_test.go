package test

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/citadelofcode/proteus/internal"
)

// Test case to validate reading of requests with bodies, query params, and keep-alive headers.
func Test_HttpRequest_ReadWithBodyAndQuery(t *testing.T) {
	server := NewTestServer(t)
	rawRequest := strings.Join([]string{
		"POST /submit/form?foo=bar&foo=baz HTTP/1.1",
		"Host: example.com",
		"Content-Length: 11",
		"Connection: keep-alive",
		"",
		"hello=world",
	}, "\r\n")

	request := NewTestRequest(t, server, strings.NewReader(rawRequest))
	err := request.Read()
	if err != nil {
		t.Fatalf(internal.TextColor.Red("Expected request to be parsed successfully, but received error: %v"), err)
	}

	if request.Method != "POST" {
		t.Errorf(internal.TextColor.Red("Expected method POST, got %s"), request.Method)
	}

	if request.ResourcePath != "/submit/form" {
		t.Errorf(internal.TextColor.Red("Expected cleaned resource path /submit/form, got %s"), request.ResourcePath)
	}

	if !bytes.Equal(request.BodyBytes, []byte("hello=world")) {
		t.Errorf(internal.TextColor.Red("Expected body bytes to match payload, got %q"), string(request.BodyBytes))
	}

	values, ok := request.Query.Get("foo")
	if !ok || len(values) != 2 || values[0] != "bar" || values[1] != "baz" {
		t.Errorf(internal.TextColor.Red("Expected query parameter foo=[bar baz], got %v"), values)
	}

	request.Locals["Started"] = time.Now().Add(-25 * time.Millisecond)
	if request.ProcessingTime() == 0 {
		t.Error(internal.TextColor.Red("ProcessingTime should return elapsed milliseconds for non-zero start time"))
	}
}

// Test case to ensure HttpRequest.Read catches malformed headers.
func Test_HttpRequest_Read_InvalidHeader(t *testing.T) {
	server := NewTestServer(t)
	rawRequest := "GET / HTTP/1.1\r\nInvalidHeaderLine\r\n\r\n"
	request := NewTestRequest(t, server, strings.NewReader(rawRequest))

	err := request.Read()
	if err == nil {
		t.Fatal(internal.TextColor.Red("Expected an error while parsing malformed header line"))
	}

	parseErr, ok := err.(*internal.RequestParseError)
	if !ok {
		t.Fatalf(internal.TextColor.Red("Expected RequestParseError, got %T"), err)
	}

	if parseErr.Section != "Header" {
		t.Errorf(internal.TextColor.Red("Expected error section Header, got %s"), parseErr.Section)
	}
}

// Test case to cover different IsConditionalGet scenarios including success, misses, and validation failures.
func Test_HttpRequest_IsConditionalGet_Scenarios(t *testing.T) {
	server := NewTestServer(t)
	root := t.TempDir()
	filePath := filepath.Join(root, "file.txt")
	err := CreateFiles(t, root, map[string][]byte{
		"file.txt": []byte("content"),
	})
	if err != nil {
		t.Fatalf(internal.TextColor.Red("Error creating temp file: %v"), err)
	}

	t.Run("Non-GET request", func(tt *testing.T) {
		request := NewTestRequest(tt, server, nil)
		request.Method = "POST"
		isCond, err := request.IsConditionalGet(filePath)
		if err != nil {
			tt.Fatalf(internal.TextColor.Red("Did not expect error: %v"), err)
		}
		if isCond {
			tt.Error(internal.TextColor.Red("Expected conditional get to be false for non-GET method"))
		}
	})

	t.Run("Invalid date header", func(tt *testing.T) {
		request := NewTestRequest(tt, server, nil)
		request.Method = "GET"
		request.Headers.Add("If-Modified-Since", "not-a-date")

		_, err := request.IsConditionalGet(filePath)
		if err == nil {
			tt.Fatal(internal.TextColor.Red("Expected error for invalid If-Modified-Since header"))
		}
	})

	t.Run("File newer than header", func(tt *testing.T) {
		request := NewTestRequest(tt, server, nil)
		request.Method = "GET"
		headerValue := time.Now().Add(-1 * time.Hour).UTC().Format(time.RFC1123)
		request.AddHeader("If-Modified-Since", headerValue)

		isCond, err := request.IsConditionalGet(filePath)
		if err != nil {
			tt.Fatalf(internal.TextColor.Red("Did not expect error: %v"), err)
		}
		if isCond {
			tt.Error(internal.TextColor.Red("Expected conditional get to be false when file is newer than header"))
		}
	})

	t.Run("File older than header", func(tt *testing.T) {
		request := NewTestRequest(tt, server, nil)
		request.Method = "GET"
		headerValue := time.Now().Add(2 * time.Hour).UTC().Format(time.RFC1123)
		request.AddHeader("If-Modified-Since", headerValue)

		isCond, err := request.IsConditionalGet(filePath)
		if err != nil {
			tt.Fatalf(internal.TextColor.Red("Did not expect error: %v"), err)
		}
		if !isCond {
			tt.Error(internal.TextColor.Red("Expected conditional get to be true when header timestamp is in the future"))
		}
	})

	t.Run("Missing file path", func(tt *testing.T) {
		request := NewTestRequest(tt, server, nil)
		request.Method = "GET"
		headerValue := time.Now().UTC().Format(time.RFC1123)
		request.AddHeader("If-Modified-Since", headerValue)

		_, err := request.IsConditionalGet(filepath.Join(root, "missing.txt"))
		if err == nil {
			tt.Error(internal.TextColor.Red("Expected error when file is missing on disk"))
		}
	})
}

// Test case to ensure invalid Content-Length header surfaces as parse error.
func Test_HttpRequest_Read_InvalidContentLength(t *testing.T) {
	server := NewTestServer(t)
	rawRequest := "POST /payload HTTP/1.1\r\nHost: example.com\r\nContent-Length: invalid\r\n\r\n"
	request := NewTestRequest(t, server, strings.NewReader(rawRequest))

	err := request.Read()
	if err == nil {
		t.Fatal(internal.TextColor.Red("Expected error for invalid Content-Length header"))
	}

	parseErr, ok := err.(*internal.RequestParseError)
	if !ok {
		t.Fatalf(internal.TextColor.Red("Expected RequestParseError, got %T"), err)
	}
	if parseErr.Section != "Header" {
		t.Errorf(internal.TextColor.Red("Expected error section Header, got %s"), parseErr.Section)
	}
}

// Test case to cover parseQueryParams failure when URL is malformed.
func Test_HttpRequest_Read_InvalidQuery(t *testing.T) {
	server := NewTestServer(t)
	rawRequest := "GET /path%ZZ HTTP/1.1\r\nHost: example.com\r\n\r\n"
	request := NewTestRequest(t, server, strings.NewReader(rawRequest))

	err := request.Read()
	if err == nil {
		t.Fatal(internal.TextColor.Red("Expected error for malformed URL"))
	}

	parseErr, ok := err.(*internal.RequestParseError)
	if !ok {
		t.Fatalf(internal.TextColor.Red("Expected RequestParseError, got %T"), err)
	}
	if parseErr.Section != "QueryParams" {
		t.Errorf(internal.TextColor.Red("Expected error section QueryParams, got %s"), parseErr.Section)
	}
}

// Test case to ensure readBody surfaces errors when payload is shorter than declared length.
func Test_HttpRequest_Read_ShortBody(t *testing.T) {
	server := NewTestServer(t)
	rawRequest := strings.Join([]string{
		"POST /short HTTP/1.1",
		"Host: example.com",
		"Content-Length: 5",
		"",
		"abc",
	}, "\r\n")
	request := NewTestRequest(t, server, strings.NewReader(rawRequest))

	err := request.Read()
	if err == nil {
		t.Fatal(internal.TextColor.Red("Expected error due to short body"))
	}

	parseErr, ok := err.(*internal.RequestParseError)
	if !ok {
		t.Fatalf(internal.TextColor.Red("Expected RequestParseError, got %T"), err)
	}
	if parseErr.Section != "Body" {
		t.Errorf(internal.TextColor.Red("Expected error section Body, got %s"), parseErr.Section)
	}
}

// Test case to ensure ProcessingTime returns zero when start time is unset.
func Test_HttpRequest_ProcessingTime_Zero(t *testing.T) {
	server := NewTestServer(t)
	req := NewTestRequest(t, server, nil)
	req.Locals["Started"] = time.Time{}

	if req.ProcessingTime() != 0 {
		t.Error(internal.TextColor.Red("Expected processing time to be zero for unset start time"))
	}
}
