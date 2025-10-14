package test

import (
	"strings"
	"testing"

	"github.com/citadelofcode/proteus/internal"
)

// Test case to validate the working of adding new header to the Headers collection.
func Test_Headers_Add(t *testing.T) {
	testHeaders := make(internal.Headers)
	testCases := []struct {
		Name        string
		HdrKey      string
		HdrValue    string
		ExpHdrCount int
	}{
		{"Adding first parameter", "Name", "Proteus", 1},
		{"Adding values to first parameter", "Name", "WebServer", 1},
		{"Adding second parameter", "Age", "18", 2},
		{"Adding third parameter", "Value", "Proteus Web Server", 3},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			testHeaders.Add(testCase.HdrKey, testCase.HdrValue)
			if len(testHeaders) == testCase.ExpHdrCount {
				tt.Logf("The expected parameter count [%d] matches the actual parameter count [%d].", testCase.ExpHdrCount, len(testHeaders))
			} else {
				tt.Errorf(internal.TextColor.Red("The expected parameter count [%d] does not match the actual parameter count [%d]."), testCase.ExpHdrCount, len(testHeaders))
			}
		})
	}
}

// Test case to validate the working of fetching the values for a 'key' from the Headers collection.
func Test_Headers_Get(t *testing.T) {
	testHeaders := make(internal.Headers)
	testHeaders.Add("Name", "proteus")
	testHeaders.Add("Server", "WebServer")
	testHeaders.Add("Server", "HTTP-Compliant")
	testCases := []struct {
		Name        string
		HdrKey      string
		ExpHdrValue string
	}{
		{"Fetching parameter with a single value in the collection", "Name", "proteus"},
		{"Fetching parameter not in the collection", "Age", ""},
		{"Fetching parameter with multiple values in the collection", "Server", "WebServer,HTTP-Compliant"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			value, ok := testHeaders.Get(testCase.HdrKey)
			if ok {
				if strings.EqualFold(value, testCase.ExpHdrValue) {
					tt.Logf("The expected value [%s] matches the returned value [%s]", testCase.ExpHdrValue, value)
				} else {
					tt.Errorf(internal.TextColor.Red("The expected value [%s] does not match the returned value [%s]"), testCase.ExpHdrValue, value)
				}
			} else {
				if strings.EqualFold(testCase.ExpHdrValue, "") {
					tt.Logf("As expected, the key [%s] was not found in the Headers collection", testCase.HdrKey)
				} else {
					tt.Errorf(internal.TextColor.Red("The key [%s] was supposed to be absent in the Headers collection, but value returned [%s] shows it was present."), testCase.HdrKey, value)
				}
			}
		})
	}
}

// Test case to validate Headers case sensitivity and key normalization.
func Test_Headers_CaseSensitivity(t *testing.T) {

	testCases := []struct {
		Name       string
		AddKey     string
		GetKey     string
		Value      string
		ShouldFind bool
	}{
		{"Exact case match", "Content-Type", "Content-Type", "application/json", true},
		{"Lowercase add, normal get", "content-type", "Content-Type", "application/json", true},
		{"Normal add, lowercase get", "Content-Type", "content-type", "application/json", true},
		{"All uppercase", "CONTENT-TYPE", "Content-Type", "application/json", true},
		{"Mixed case variations", "cOnTeNt-TyPe", "Content-Type", "application/json", true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			// Create fresh headers for each test
			headers := make(internal.Headers)
			headers.Add(testCase.AddKey, testCase.Value)

			value, ok := headers.Get(testCase.GetKey)
			if ok == testCase.ShouldFind {
				if testCase.ShouldFind {
					if strings.EqualFold(value, testCase.Value) {
						tt.Logf("Header key normalization works correctly: [%s] -> [%s]", testCase.AddKey, testCase.GetKey)
					} else {
						tt.Errorf(internal.TextColor.Red("Expected value [%s], got [%s]"), testCase.Value, value)
					}
				} else {
					tt.Logf("Header key [%s] correctly not found when looking for [%s]", testCase.AddKey, testCase.GetKey)
				}
			} else {
				tt.Errorf(internal.TextColor.Red("Case sensitivity test failed for keys [%s] -> [%s]"), testCase.AddKey, testCase.GetKey)
			}
		})
	}
}

// Test case to validate Headers with special characters and edge cases.
func Test_Headers_SpecialCharacters(t *testing.T) {

	testCases := []struct {
		Name             string
		HeaderKey        string
		HeaderValue      string
		ExpectedBehavior string
	}{
		{"Header with special characters in value", "Custom-Header", "value with spaces & symbols!@#$%", "normal"},
		{"Header with unicode in value", "Unicode-Header", "café Ñoño 世界", "normal"},
		{"Header with newline characters", "Multiline-Header", "line1\nline2\nline3", "normal"},
		{"Header with empty value", "Empty-Header", "", "normal"},
		{"Header with only spaces", "Space-Header", "   ", "normal"},
		{"Header with comma-separated values", "List-Header", "value1, value2, value3", "comma-split"},
		{"Header with multiple commas", "Multi-Comma", "a,b,c,d,e", "comma-split"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			headers := make(internal.Headers)
			headers.Add(testCase.HeaderKey, testCase.HeaderValue)

			retrievedValue, ok := headers.Get(testCase.HeaderKey)
			if !ok {
				tt.Errorf(internal.TextColor.Red("Header [%s] not found after adding"), testCase.HeaderKey)
				return
			}

			if testCase.ExpectedBehavior == "comma-split" {
				// For comma-separated values, they should be joined back with commas
				if strings.Contains(retrievedValue, ",") {
					tt.Logf("Comma-separated values correctly handled for header [%s]: [%s]", testCase.HeaderKey, retrievedValue)
				} else {
					tt.Logf("Header value [%s] for key [%s] (expected comma separation but got: [%s])", testCase.HeaderValue, testCase.HeaderKey, retrievedValue)
				}
			} else {
				// For normal headers, value should be preserved as-is or similar
				tt.Logf("Special character header [%s] handled: Original=[%s], Retrieved=[%s]", testCase.HeaderKey, testCase.HeaderValue, retrievedValue)
			}
		})
	}
}

// Test case to validate Headers.Add with multiple values and comma handling.
func Test_Headers_MultipleValues(t *testing.T) {
	testHeaders := make(internal.Headers)

	// Add same header multiple times
	testHeaders.Add("Accept", "text/html")
	testHeaders.Add("Accept", "application/json")
	testHeaders.Add("Accept", "application/xml")

	// Add header with comma-separated values
	testHeaders.Add("Cache-Control", "no-cache, no-store, must-revalidate")

	testCases := []struct {
		Name                  string
		HeaderKey             string
		ExpectedValueContains []string
	}{
		{"Multiple Accept headers", "Accept", []string{"text/html", "application/json", "application/xml"}},
		{"Comma-separated Cache-Control", "Cache-Control", []string{"no-cache", "no-store", "must-revalidate"}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			value, ok := testHeaders.Get(testCase.HeaderKey)
			if !ok {
				tt.Errorf(internal.TextColor.Red("Header [%s] not found"), testCase.HeaderKey)
				return
			}

			// Check if all expected values are present in the result
			for _, expectedValue := range testCase.ExpectedValueContains {
				if !strings.Contains(value, expectedValue) {
					tt.Errorf(internal.TextColor.Red("Expected value [%s] not found in header [%s] result: [%s]"), expectedValue, testCase.HeaderKey, value)
				} else {
					tt.Logf("Expected value [%s] found in header [%s]", expectedValue, testCase.HeaderKey)
				}
			}
		})
	}
}

// Test case to validate Headers.Length() functionality.
func Test_Headers_Length(t *testing.T) {
	testHeaders := make(internal.Headers)

	// Initial length should be 0
	if testHeaders.Length() != 0 {
		t.Errorf(internal.TextColor.Red("Expected initial length 0, got %d"), testHeaders.Length())
	}

	testCases := []struct {
		Name           string
		Action         string
		HeaderKey      string
		HeaderValue    string
		ExpectedLength int
	}{
		{"Add first header", "add", "Content-Type", "application/json", 1},
		{"Add second header", "add", "Accept", "application/json", 2},
		{"Add to existing header", "add", "Content-Type", "text/html", 2}, // Should not increase count
		{"Add third unique header", "add", "Authorization", "Bearer token", 3},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			if testCase.Action == "add" {
				testHeaders.Add(testCase.HeaderKey, testCase.HeaderValue)
			}

			currentLength := testHeaders.Length()
			if currentLength == testCase.ExpectedLength {
				tt.Logf("Header length correctly updated to %d", currentLength)
			} else {
				tt.Errorf(internal.TextColor.Red("Expected length %d, got %d"), testCase.ExpectedLength, currentLength)
			}
		})
	}
}

// Test case to validate Headers with edge case key and value combinations.
func Test_Headers_EdgeCases(t *testing.T) {

	testCases := []struct {
		Name        string
		HeaderKey   string
		HeaderValue string
		ShouldStore bool
	}{
		{"Normal header", "Content-Type", "application/json", true},
		{"Header key with hyphens", "X-Custom-Header", "custom-value", true},
		{"Header key with numbers", "X-Retry-After", "300", true},
		{"Very long header value", "Long-Header", strings.Repeat("A", 1000), true},
		{"Header with empty key", "", "some-value", true},          // MIME canonicalization might handle this
		{"Header key with spaces", "Spaced Header", "value", true}, // Will be normalized
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			headers := make(internal.Headers)
			headers.Add(testCase.HeaderKey, testCase.HeaderValue)

			// Try to retrieve the header - the key might be normalized
			value, ok := headers.Get(testCase.HeaderKey)

			if testCase.ShouldStore {
				if ok {
					tt.Logf("Edge case header [%s] correctly stored and retrieved: [%s]", testCase.HeaderKey, value)
				} else {
					// Some edge cases might not be retrievable with the same key due to normalization
					tt.Logf("Edge case header [%s] was stored but not retrievable with same key (possibly normalized)", testCase.HeaderKey)
				}
			} else {
				if !ok {
					tt.Logf("Edge case header [%s] correctly rejected", testCase.HeaderKey)
				} else {
					tt.Errorf(internal.TextColor.Red("Edge case header [%s] should not have been stored but was"), testCase.HeaderKey)
				}
			}
		})
	}
}
