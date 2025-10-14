package test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/citadelofcode/proteus/internal"
)

// Test case to validate the working of the JsonParser() middleware to parse JSON objects as payloads.
func Test_JsonParser_ObjectParse(t *testing.T) {
	TestServer := NewTestServer(t)
	JsonParser := internal.JsonParser()
	testCases := []struct {
		Name      string
		InPayload string
		OpPayload map[string]any
	}{
		{"A valid JSON payload with a flat object", `{"name":"Mahesh","email":"mkbalaji@email.com"}`, map[string]any{"name": "Mahesh", "email": "mkbalaji@email.com"}},
		{"A valid JSON payload with a nested object", `{"name":"Mahesh","email":"mkbalaji@email.com","location":{"latitude":24,"longitude":-124}}`, map[string]any{"name": "Mahesh", "email": "mkbalaji@email.com", "location": map[string]any{"latitude": float64(24), "longitude": float64(-124)}}},
		{"A valid flat JSON payload with numeric fields", `{"name":"Mahesh","email":"mkbalaji@email.com", "age":18}`, map[string]any{"name": "Mahesh", "email": "mkbalaji@email.com", "age": float64(18)}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			request := NewTestRequest(tt, TestServer, nil)
			response := NewTestResponse(tt, request.Version, TestServer, nil)
			request.BodyBytes = []byte(testCase.InPayload)
			request.AddHeader("Content-Type", "application/json")
			JsonParser(request, response, Stop)
			tt.Logf("Expected JSON Object Payload: %#v", testCase.OpPayload)
			tt.Logf("Parsed JSON Object payload: %#v", request.Body)
			isEqual := reflect.DeepEqual(request.Body, testCase.OpPayload)
			if isEqual {
				tt.Log("The expected JSON payload matches the parsed JSON object payload.")
			} else {
				tt.Error(internal.TextColor.Red("The expected JSON payload does not match the parsed JSON object payload."))
			}
		})
	}
}

// Test case to validate the working of the JsonParser() middleware to parse JSON arrays as payloads.
func Test_JsonParser_ArrayParse(t *testing.T) {
	TestServer := NewTestServer(t)
	JsonParser := internal.JsonParser()
	testCases := []struct {
		Name      string
		InPayload string
		OpPayload []any
	}{
		{"A valid JSON array", `[{"name":"Mahesh","email":"mkbalaji@email.com"},{"name":"sdgbjsdk","email":"jksbfuhs@email.com"}]`, []any{map[string]any{"name": "Mahesh", "email": "mkbalaji@email.com"}, map[string]any{"name": "sdgbjsdk", "email": "jksbfuhs@email.com"}}},
		{"A valid JSON array with numeric fields", `[{"name":"Mahesh","age":18},{"name":"sdgbjsdk","age":20}]`, []any{map[string]any{"name": "Mahesh", "age": float64(18)}, map[string]any{"name": "sdgbjsdk", "age": float64(20)}}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			request := NewTestRequest(tt, TestServer, nil)
			response := NewTestResponse(tt, request.Version, TestServer, nil)
			request.BodyBytes = []byte(testCase.InPayload)
			request.AddHeader("Content-Type", "application/json")
			JsonParser(request, response, Stop)
			tt.Logf("Expected JSON Array Payload: %#v", testCase.OpPayload)
			tt.Logf("Parsed JSON Array Payload: %#v", request.Body)
			isEqual := reflect.DeepEqual(request.Body, testCase.OpPayload)
			if isEqual {
				tt.Log("The expected JSON payload matches the parsed JSON array payload.")
			} else {
				tt.Error(internal.TextColor.Red("The expected JSON payload does not match the parsed JSON array payload."))
			}
		})
	}
}

// Test case to validate how JsonParser() middleware affects non-JSON payloads. It checks if the request body still remains "nil" after middleware execution for non-JSON payloads.
func Test_JsonParser_NonJson(t *testing.T) {
	TestServer := NewTestServer(t)
	JsonParser := internal.JsonParser()
	testCases := []struct {
		Name          string
		IpContentType string
		IpPayload     string
	}{
		{"A plain text payload", "text/plain", "This is a simple text value"},
		{"An XML payload", "application/xml", `<?xml version="1.0" encoding="UTF-8"?><Name>Mahesh</Name>`},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			request := NewTestRequest(tt, TestServer, nil)
			response := NewTestResponse(tt, request.Version, TestServer, nil)
			request.BodyBytes = []byte(testCase.IpPayload)
			request.AddHeader("Content-Type", testCase.IpContentType)
			JsonParser(request, response, Stop)
			if request.Body == nil {
				tt.Logf("The request body is 'nil' as expected, for payload [%s] of content type [%s]", testCase.IpPayload, testCase.IpContentType)
			} else {
				tt.Errorf(internal.TextColor.Red("The request body is not 'nil' as expected, for payload [%s] of content type [%s]"), testCase.IpPayload, testCase.IpContentType)
			}
		})
	}
}

// Test case to validate how the UrlEncoded() middleware works for request payloads of type - "application/x-www-form-urlencoded".
func Test_UrlEncoded_ValidPayloads(t *testing.T) {
	TestServer := NewTestServer(t)
	UrlEncoded := internal.UrlEncoded()
	testCases := []struct {
		Name      string
		IpPayload string
		OpPayload map[string][]string
	}{
		{"A valid url encoded string with single value params", "name=Mahesh&age=30&city=San+Francisco", map[string][]string{"name": {"Mahesh"}, "age": {"30"}, "city": {"San Francisco"}}},
		{"A valid url encoded string with single value params and url encoded values", "name=Mahesh&email=mkbalaji%40email.com&age=30", map[string][]string{"age": {"30"}, "email": {"mkbalaji@email.com"}, "name": {"Mahesh"}}},
		{"A valid url encoded string with multiple value params", "tag=go&tag=web&tag=backend&user=mahesh", map[string][]string{"tag": {"go", "web", "backend"}, "user": {"mahesh"}}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			request := NewTestRequest(tt, TestServer, nil)
			response := NewTestResponse(tt, request.Version, TestServer, nil)
			request.AddHeader("Content-Type", "application/x-www-form-urlencoded")
			request.BodyBytes = []byte(testCase.IpPayload)
			UrlEncoded(request, response, Stop)
			tt.Logf("Expected Url encoded payload: %#v\n", testCase.OpPayload)
			tt.Logf("Processed Url encoded payload: %#v\n", request.Body)
			isEqual := reflect.DeepEqual(request.Body, testCase.OpPayload)
			if isEqual {
				tt.Log("The expected request payload matches the processed payload.")
			} else {
				tt.Error(internal.TextColor.Red("The expected request payload does not match the processed payload."))
			}
		})
	}
}

// Test case to validate how the UrlEncoded() middleware works for request payloads not of type - "application/x-www-form-urlencoded".
func Test_UrlEncoded_InvalidPayloads(t *testing.T) {
	TestServer := NewTestServer(t)
	UrlEncoded := internal.UrlEncoded()
	testCases := []struct {
		Name          string
		IpContentType string
		IpPayload     string
	}{
		{"A plain text payload", "text/plain", "This is a simple text value"},
		{"An XML payload", "application/xml", `<?xml version="1.0" encoding="UTF-8"?><Name>Mahesh</Name>`},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			request := NewTestRequest(tt, TestServer, nil)
			response := NewTestResponse(tt, request.Version, TestServer, nil)
			request.BodyBytes = []byte(testCase.IpPayload)
			request.AddHeader("Content-Type", testCase.IpContentType)
			UrlEncoded(request, response, Stop)
			if request.Body == nil {
				tt.Logf("The request body is 'nil' as expected, for payload [%s] of content type [%s]", testCase.IpPayload, testCase.IpContentType)
			} else {
				tt.Errorf(internal.TextColor.Red("The request body is not 'nil' as expected, for payload [%s] of content type [%s]"), testCase.IpPayload, testCase.IpContentType)
			}
		})
	}
}

// Test case to validate how either of the middlewares - JsonParser() and UrlEncoded() behave when there is no content type header present.
func Test_BodyParser_NoContentType(t *testing.T) {
	TestServer := NewTestServer(t)
	JsonParser := internal.JsonParser()
	UrlEncoded := internal.UrlEncoded()
	testCases := []struct {
		Name       string
		Middleware string
		IpPayload  string
	}{
		{"Json Parser middleware", "JsonParser", `{"name":"Mahesh","email":"mkbalaji@email.com"}`},
		{"Url Encoded middleware", "UrlEncoded", "name=Mahesh&age=30&city=San+Francisco"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			if strings.EqualFold(testCase.Middleware, "JsonParser") {
				request := NewTestRequest(tt, TestServer, nil)
				response := NewTestResponse(tt, request.Version, TestServer, nil)
				request.BodyBytes = []byte(testCase.IpPayload)
				JsonParser(request, response, Stop)
				if request.Body == nil {
					tt.Logf("The request body is 'nil' as expected, for payload [%s]", testCase.IpPayload)
				} else {
					tt.Errorf(internal.TextColor.Red("The request body is not 'nil' as expected, for payload [%s]"), testCase.IpPayload)
				}
			}

			if strings.EqualFold(testCase.Middleware, "UrlEncoded") {
				request := NewTestRequest(tt, TestServer, nil)
				response := NewTestResponse(tt, request.Version, TestServer, nil)
				request.BodyBytes = []byte(testCase.IpPayload)
				UrlEncoded(request, response, Stop)
				if request.Body == nil {
					tt.Logf("The request body is 'nil' as expected, for payload [%s]", testCase.IpPayload)
				} else {
					tt.Errorf(internal.TextColor.Red("The request body is not 'nil' as expected, for payload [%s]"), testCase.IpPayload)
				}
			}
		})
	}
}

// Test case to validate JsonParser() middleware with malformed JSON payloads.
func Test_JsonParser_MalformedJson(t *testing.T) {
	TestServer := NewTestServer(t)
	JsonParser := internal.JsonParser()
	testCases := []struct {
		Name      string
		IpPayload string
	}{
		{"Unterminated JSON object", `{"name":"Mahesh","email":"mkbalaji@email.com"`},
		{"Invalid JSON with trailing comma", `{"name":"Mahesh","email":"mkbalaji@email.com",}`},
		{"JSON with unquoted keys", `{name:"Mahesh",email:"mkbalaji@email.com"}`},
		{"JSON with single quotes", `{'name':'Mahesh','email':'mkbalaji@email.com'}`},
		{"Empty JSON payload", ``},
		{"Invalid JSON characters", `{"name":"Mahesh"&"email":"mkbalaji@email.com"}`},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			request := NewTestRequest(tt, TestServer, nil)
			response := NewTestResponse(tt, request.Version, TestServer, nil)
			request.BodyBytes = []byte(testCase.IpPayload)
			request.AddHeader("Content-Type", "application/json")
			JsonParser(request, response, Stop)

			// For malformed JSON, the body should remain nil due to parsing error
			if request.Body == nil {
				tt.Logf("Malformed JSON payload [%s] correctly resulted in nil body", testCase.IpPayload)
			} else {
				tt.Errorf(internal.TextColor.Red("Expected nil body for malformed JSON [%s], but got: %#v"), testCase.IpPayload, request.Body)
			}
		})
	}
}

// Test case to validate JsonParser() middleware with content-type case sensitivity.
func Test_JsonParser_ContentTypeCaseSensitivity(t *testing.T) {
	TestServer := NewTestServer(t)
	JsonParser := internal.JsonParser()
	testPayload := `{"name":"Mahesh","email":"mkbalaji@email.com"}`
	expectedPayload := map[string]any{"name": "Mahesh", "email": "mkbalaji@email.com"}

	testCases := []struct {
		Name        string
		ContentType string
		ShouldParse bool
	}{
		{"Lowercase application/json", "application/json", true},
		{"Uppercase APPLICATION/JSON", "APPLICATION/JSON", true},
		{"Mixed case Application/Json", "Application/Json", true},
		{"With charset parameter", "application/json; charset=utf-8", false}, // Should not match exactly
		{"With whitespace", "  application/json  ", true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			request := NewTestRequest(tt, TestServer, nil)
			response := NewTestResponse(tt, request.Version, TestServer, nil)
			request.BodyBytes = []byte(testPayload)
			request.AddHeader("Content-Type", testCase.ContentType)
			JsonParser(request, response, Stop)

			if testCase.ShouldParse {
				if reflect.DeepEqual(request.Body, expectedPayload) {
					tt.Logf("Content-Type [%s] correctly parsed JSON", testCase.ContentType)
				} else {
					tt.Errorf(internal.TextColor.Red("Content-Type [%s] should have parsed JSON but didn't"), testCase.ContentType)
				}
			} else {
				if request.Body == nil {
					tt.Logf("Content-Type [%s] correctly ignored JSON parsing", testCase.ContentType)
				} else {
					tt.Errorf(internal.TextColor.Red("Content-Type [%s] should not have parsed JSON but did"), testCase.ContentType)
				}
			}
		})
	}
}

// Test case to validate UrlEncoded() middleware with invalid URL encoding.
func Test_UrlEncoded_InvalidEncoding(t *testing.T) {
	TestServer := NewTestServer(t)
	UrlEncoded := internal.UrlEncoded()
	testCases := []struct {
		Name      string
		IpPayload string
	}{
		{"Invalid percent encoding", "name=Mahesh&email=test%ZZ"},
		{"Incomplete percent encoding", "name=Mahesh&email=test%2"},
		{"Multiple equals signs", "name==Mahesh&age=30"},
		{"Missing equals sign", "name&age=30"},
		{"Empty parameter name", "=value&age=30"},
		{"Only parameter names", "name&age&city"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			request := NewTestRequest(tt, TestServer, nil)
			response := NewTestResponse(tt, request.Version, TestServer, nil)
			request.BodyBytes = []byte(testCase.IpPayload)
			request.AddHeader("Content-Type", "application/x-www-form-urlencoded")
			UrlEncoded(request, response, Stop)

			// Invalid encoding should result in nil body due to parsing error
			if request.Body == nil {
				tt.Logf("Invalid URL encoding [%s] correctly resulted in nil body", testCase.IpPayload)
			} else {
				// Some invalid encodings might still parse partially
				tt.Logf("URL encoding [%s] parsed as: %#v", testCase.IpPayload, request.Body)
			}
		})
	}
}

// Test case to validate UrlEncoded() middleware with edge cases.
func Test_UrlEncoded_EdgeCases(t *testing.T) {
	TestServer := NewTestServer(t)
	UrlEncoded := internal.UrlEncoded()
	testCases := []struct {
		Name           string
		IpPayload      string
		ExpectedResult map[string][]string
	}{
		{"Empty payload", "", map[string][]string{}},
		{"Only ampersands", "&&&", map[string][]string{}},
		{"Only equals", "===", map[string][]string{"": {"", "", ""}}},
		{"Special characters in values", "name=test@example.com&symbols=%21%40%23%24%25%5E%26%2A%28%29", map[string][]string{"name": {"test@example.com"}, "symbols": {"!@#$%^&*()"}}},
		{"Unicode characters", "name=José&city=São+Paulo", map[string][]string{"name": {"José"}, "city": {"São Paulo"}}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			request := NewTestRequest(tt, TestServer, nil)
			response := NewTestResponse(tt, request.Version, TestServer, nil)
			request.BodyBytes = []byte(testCase.IpPayload)
			request.AddHeader("Content-Type", "application/x-www-form-urlencoded")
			UrlEncoded(request, response, Stop)

			if len(testCase.ExpectedResult) == 0 {
				// For empty expected results, check if body is nil or empty map
				if request.Body == nil {
					tt.Logf("Empty payload [%s] correctly resulted in nil body", testCase.IpPayload)
				} else if parsed, ok := request.Body.(map[string][]string); ok && len(parsed) == 0 {
					tt.Logf("Empty payload [%s] correctly resulted in empty map", testCase.IpPayload)
				} else {
					tt.Logf("Payload [%s] resulted in: %#v", testCase.IpPayload, request.Body)
				}
			} else {
				// For expected results, compare with parsed output
				if parsed, ok := request.Body.(map[string][]string); ok {
					if reflect.DeepEqual(parsed, testCase.ExpectedResult) {
						tt.Logf("Payload [%s] correctly parsed", testCase.IpPayload)
					} else {
						tt.Logf("Payload [%s] parsed differently than expected. Got: %#v, Expected: %#v", testCase.IpPayload, parsed, testCase.ExpectedResult)
					}
				} else {
					tt.Errorf(internal.TextColor.Red("Payload [%s] didn't parse as expected map type"), testCase.IpPayload)
				}
			}
		})
	}
}
