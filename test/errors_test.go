package test

import (
	"strings"
	"testing"
	"github.com/citadelofcode/proteus/internal"
)

// Test case to validate RequestParseError functionality.
func Test_RequestParseError(t *testing.T) {
	testCases := []struct {
		Name string
		Section string
		Value string
		Message string
		ExpectedErrorMessage string
	} {
		{ "Header parsing error", "Header", "Invalid-Header: value", "Malformed header format", "RequestParseError :: Section: (Header) :: Value: (Invalid-Header: value) :: Malformed header format" },
		{ "Body parsing error", "Body", `{"invalid": json}`, "Unexpected character in JSON", "RequestParseError :: Section: (Body) :: Value: ({\"invalid\": json}) :: Unexpected character in JSON" },
		{ "Query params parsing error", "QueryParams", "name=&age=", "Empty parameter values", "RequestParseError :: Section: (QueryParams) :: Value: (name=&age=) :: Empty parameter values" },
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			rpe := &internal.RequestParseError{
				Section: testCase.Section,
				Value: testCase.Value,
				Message: testCase.Message,
			}

			errorMessage := rpe.Error()
			if strings.EqualFold(errorMessage, testCase.ExpectedErrorMessage) {
				tt.Logf("RequestParseError message matches expected format: %s", errorMessage)
			} else {
				tt.Errorf(internal.TextColor.Red("Expected error message [%s], but got [%s]"), testCase.ExpectedErrorMessage, errorMessage)
			}

			// Test that it implements the error interface
			var err error = rpe
			if err == nil {
				tt.Error(internal.TextColor.Red("RequestParseError should implement the error interface"))
			}
		})
	}
}

// Test case to validate RoutingError functionality.
func Test_RoutingError(t *testing.T) {
	testCases := []struct {
		Name string
		RoutePath string
		Message string
		ExpectedErrorMessage string
	} {
		{ "Route not found error", "/api/users/123", "No matching route found", "RoutingError :: Route - [/api/users/123] :: No matching route found" },
		{ "Method not allowed error", "/api/posts", "POST method not allowed for this route", "RoutingError :: Route - [/api/posts] :: POST method not allowed for this route" },
		{ "Invalid route pattern", "/users/:id/", "Invalid route pattern with trailing slash", "RoutingError :: Route - [/users/:id/] :: Invalid route pattern with trailing slash" },
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			re := &internal.RoutingError{
				RoutePath: testCase.RoutePath,
				Message: testCase.Message,
			}

			errorMessage := re.Error()
			if strings.EqualFold(errorMessage, testCase.ExpectedErrorMessage) {
				tt.Logf("RoutingError message matches expected format: %s", errorMessage)
			} else {
				tt.Errorf(internal.TextColor.Red("Expected error message [%s], but got [%s]"), testCase.ExpectedErrorMessage, errorMessage)
			}

			// Test that it implements the error interface
			var err error = re
			if err == nil {
				tt.Error(internal.TextColor.Red("RoutingError should implement the error interface"))
			}
		})
	}
}

// Test case to validate ResponseError functionality.
func Test_ResponseError(t *testing.T) {
	testCases := []struct {
		Name string
		Section string
		Value string
		Message string
		ExpectedErrorMessage string
	} {
		{ "Header writing error", "Header", "Content-Type: invalid/type", "Invalid content type format", "ResponseError :: Section: (Header) :: Value: (Content-Type: invalid/type) :: Invalid content type format" },
		{ "Body writing error", "Body", "Response body content", "Failed to write response body", "ResponseError :: Section: (Body) :: Value: (Response body content) :: Failed to write response body" },
		{ "Status line error", "StatusLine", "HTTP/1.1 200", "Missing status message", "ResponseError :: Section: (StatusLine) :: Value: (HTTP/1.1 200) :: Missing status message" },
		{ "Response write error", "RespWrite", "Connection closed", "Connection was closed before response completed", "ResponseError :: Section: (RespWrite) :: Value: (Connection closed) :: Connection was closed before response completed" },
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			resErr := internal.ResponseError{
				Section: testCase.Section,
				Value: testCase.Value,
				Message: testCase.Message,
			}

			errorMessage := resErr.Error()
			if strings.EqualFold(errorMessage, testCase.ExpectedErrorMessage) {
				tt.Logf("ResponseError message matches expected format: %s", errorMessage)
			} else {
				tt.Errorf(internal.TextColor.Red("Expected error message [%s], but got [%s]"), testCase.ExpectedErrorMessage, errorMessage)
			}

			// Test that it implements the error interface
			var err error = resErr
			if err.Error() != errorMessage {
				tt.Error(internal.TextColor.Red("ResponseError should implement the error interface correctly"))
			}
		})
	}
}

// Test case to validate ReadTimeoutError functionality.
func Test_ReadTimeoutError(t *testing.T) {
	rte := &internal.ReadTimeoutError{}
	expectedMessage := "Read timeout error occurred on the underlying TCP Connection."
	
	errorMessage := rte.Error()
	if strings.EqualFold(errorMessage, expectedMessage) {
		t.Logf("ReadTimeoutError message matches expected format: %s", errorMessage)
	} else {
		t.Errorf(internal.TextColor.Red("Expected error message [%s], but got [%s]"), expectedMessage, errorMessage)
	}

	// Test that it implements the error interface
	var err error = rte
	if err == nil {
		t.Error(internal.TextColor.Red("ReadTimeoutError should implement the error interface"))
	}
}

// Test case to validate FileSystemError functionality.
func Test_FileSystemError(t *testing.T) {
	testCases := []struct {
		Name string
		TargetPath string
		Message string
		ExpectedErrorMessage string
	} {
		{ "File not found error", "/path/to/nonexistent/file.txt", "File does not exist", "File System Error for [/path/to/nonexistent/file.txt] :: File does not exist" },
		{ "Permission denied error", "/root/protected/file.txt", "Permission denied", "File System Error for [/root/protected/file.txt] :: Permission denied" },
		{ "Directory expected error", "/path/to/file.txt", "Expected directory but found file", "File System Error for [/path/to/file.txt] :: Expected directory but found file" },
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			fsErr := &internal.FileSystemError{
				TargetPath: testCase.TargetPath,
				Message: testCase.Message,
			}

			errorMessage := fsErr.Error()
			if strings.EqualFold(errorMessage, testCase.ExpectedErrorMessage) {
				tt.Logf("FileSystemError message matches expected format: %s", errorMessage)
			} else {
				tt.Errorf(internal.TextColor.Red("Expected error message [%s], but got [%s]"), testCase.ExpectedErrorMessage, errorMessage)
			}

			// Test that it implements the error interface
			var err error = fsErr
			if err == nil {
				tt.Error(internal.TextColor.Red("FileSystemError should implement the error interface"))
			}
		})
	}
}

// Test case to validate CustomError functionality.
func Test_CustomError(t *testing.T) {
	testCases := []struct {
		Name string
		Message string
		ExpectedErrorMessage string
	} {
		{ "Authentication error", "Invalid credentials provided", "Request Response Error :: Invalid credentials provided" },
		{ "Validation error", "Required field 'email' is missing", "Request Response Error :: Required field 'email' is missing" },
		{ "Business logic error", "User account is already activated", "Request Response Error :: User account is already activated" },
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			ce := &internal.CustomError{
				Message: testCase.Message,
			}

			errorMessage := ce.Error()
			if strings.EqualFold(errorMessage, testCase.ExpectedErrorMessage) {
				tt.Logf("CustomError message matches expected format: %s", errorMessage)
			} else {
				tt.Errorf(internal.TextColor.Red("Expected error message [%s], but got [%s]"), testCase.ExpectedErrorMessage, errorMessage)
			}

			// Test that it implements the error interface
			var err error = ce
			if err == nil {
				tt.Error(internal.TextColor.Red("CustomError should implement the error interface"))
			}
		})
	}
}

// Test case to validate that all custom errors can be differentiated by type assertion.
func Test_ErrorTypes_TypeAssertion(t *testing.T) {
	errors := []struct {
		Name string
		Error error
		ExpectedType string
	} {
		{ "RequestParseError type assertion", &internal.RequestParseError{Section: "Test", Value: "test", Message: "test"}, "RequestParseError" },
		{ "RoutingError type assertion", &internal.RoutingError{RoutePath: "/test", Message: "test"}, "RoutingError" },
		{ "ResponseError type assertion", internal.ResponseError{Section: "Test", Value: "test", Message: "test"}, "ResponseError" },
		{ "ReadTimeoutError type assertion", &internal.ReadTimeoutError{}, "ReadTimeoutError" },
		{ "FileSystemError type assertion", &internal.FileSystemError{TargetPath: "/test", Message: "test"}, "FileSystemError" },
		{ "CustomError type assertion", &internal.CustomError{Message: "test"}, "CustomError" },
	}

	for _, testCase := range errors {
		t.Run(testCase.Name, func(tt *testing.T) {
			switch testCase.ExpectedType {
			case "RequestParseError":
				if _, ok := testCase.Error.(*internal.RequestParseError); !ok {
					tt.Errorf(internal.TextColor.Red("Expected RequestParseError type, but got %T"), testCase.Error)
				} else {
					tt.Log("RequestParseError type assertion successful")
				}
			case "RoutingError":
				if _, ok := testCase.Error.(*internal.RoutingError); !ok {
					tt.Errorf(internal.TextColor.Red("Expected RoutingError type, but got %T"), testCase.Error)
				} else {
					tt.Log("RoutingError type assertion successful")
				}
			case "ResponseError":
				if _, ok := testCase.Error.(internal.ResponseError); !ok {
					tt.Errorf(internal.TextColor.Red("Expected ResponseError type, but got %T"), testCase.Error)
				} else {
					tt.Log("ResponseError type assertion successful")
				}
			case "ReadTimeoutError":
				if _, ok := testCase.Error.(*internal.ReadTimeoutError); !ok {
					tt.Errorf(internal.TextColor.Red("Expected ReadTimeoutError type, but got %T"), testCase.Error)
				} else {
					tt.Log("ReadTimeoutError type assertion successful")
				}
			case "FileSystemError":
				if _, ok := testCase.Error.(*internal.FileSystemError); !ok {
					tt.Errorf(internal.TextColor.Red("Expected FileSystemError type, but got %T"), testCase.Error)
				} else {
					tt.Log("FileSystemError type assertion successful")
				}
			case "CustomError":
				if _, ok := testCase.Error.(*internal.CustomError); !ok {
					tt.Errorf(internal.TextColor.Red("Expected CustomError type, but got %T"), testCase.Error)
				} else {
					tt.Log("CustomError type assertion successful")
				}
			}
		})
	}
}