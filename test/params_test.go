package test

import (
	"slices"
	"testing"

	"github.com/citadelofcode/proteus/internal"
)

// Test case to validate the working of adding new parameter to the Params collection.
func Test_Params_Add(t *testing.T) {
	testParams := make(internal.Params)
	testCases := []struct {
		Name          string
		ParamKey      string
		ParamValues   []string
		ExpParamCount int
	}{
		{"Adding first parameter", "Name", []string{"Proteus"}, 1},
		{"Adding values to first parameter", "Name", []string{"WebServer"}, 1},
		{"Adding second parameter", "Age", []string{"18"}, 2},
		{"Adding third parameter", "Value", []string{"Proteus Web Server"}, 3},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			testParams.Add(testCase.ParamKey, testCase.ParamValues)
			if len(testParams) == testCase.ExpParamCount {
				tt.Logf("The expected parameter count [%d] matches the actual parameter count [%d].", testCase.ExpParamCount, len(testParams))
			} else {
				tt.Errorf(internal.TextColor.Red("The expected parameter count [%d] does not match the actual parameter count [%d]."), testCase.ExpParamCount, len(testParams))
			}
		})
	}
}

// Test case to validate the working of fetching the values for a 'key' from the params collection.
func Test_Params_Get(t *testing.T) {
	testParams := make(internal.Params)
	testParams.Add("Name", []string{"proteus"})
	testCases := []struct {
		Name           string
		ParamKey       string
		ExpParamValues []string
	}{
		{"Fetching parameter in the collection", "Name", []string{"proteus"}},
		{"Fetching parameter not in the collection", "Age", nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			values, _ := testParams.Get(testCase.ParamKey)
			if testCase.ExpParamValues == nil {
				if values != nil {
					tt.Errorf(internal.TextColor.Red("Expected returned values array to be nil, but got a non-nil value instead - %#v"), values)
				} else {
					tt.Logf("Expected a nil slice, and got exactly that.")
				}

				return
			}

			if slices.Equal(testCase.ExpParamValues, values) {
				tt.Logf("The returned slice of values [%v], matches the expected slice of values [%v]", values, testCase.ExpParamValues)
			} else {
				tt.Errorf(internal.TextColor.Red("The returned slice of values [%v], does not match the expected slice of values [%#v]"), values, testCase.ExpParamValues)
			}
		})
	}
}

// Test case to validate Params with empty and whitespace parameter values.
func Test_Params_EmptyValues(t *testing.T) {
	testCases := []struct {
		Name           string
		ParamKey       string
		ParamValues    []string
		ExpectedLength int
	}{
		{"Empty string values", "empty", []string{""}, 1},
		{"Multiple empty values", "multi-empty", []string{"", "", ""}, 1},
		{"Whitespace only values", "whitespace", []string{"   ", "\t", "\n"}, 1},
		{"Mixed empty and non-empty", "mixed", []string{"", "value", "", "another"}, 1},
		{"Empty slice", "empty-slice", []string{}, 1},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			params := make(internal.Params)
			params.Add(testCase.ParamKey, testCase.ParamValues)

			if params.Length() == testCase.ExpectedLength {
				tt.Logf("Parameter count correctly updated to %d", params.Length())
			} else {
				tt.Errorf(internal.TextColor.Red("Expected parameter count %d, got %d"), testCase.ExpectedLength, params.Length())
			}

			// Verify we can retrieve the values
			retrievedValues, ok := params.Get(testCase.ParamKey)
			if !ok {
				tt.Errorf(internal.TextColor.Red("Parameter [%s] not found after adding"), testCase.ParamKey)
			} else {
				tt.Logf("Parameter [%s] stored values: %#v", testCase.ParamKey, retrievedValues)
			}
		})
	}
}

// Test case to validate Params with special characters and URL encoding.
func Test_Params_SpecialCharacters(t *testing.T) {
	testCases := []struct {
		Name        string
		ParamKey    string
		ParamValues []string
	}{
		{"Unicode characters", "unicode", []string{"café", "Ñoño", "世界"}},
		{"Special symbols", "symbols", []string{"!@#$%^&*()", "<>&\"'"}},
		{"URL encoded characters", "encoded", []string{"%20", "%2B", "%3D"}},
		{"Mixed special characters", "mixed", []string{"hello world", "test@example.com", "value=123"}},
		{"Newlines and tabs", "whitespace", []string{"line1\nline2", "col1\tcol2"}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			params := make(internal.Params)
			params.Add(testCase.ParamKey, testCase.ParamValues)

			retrievedValues, ok := params.Get(testCase.ParamKey)
			if !ok {
				tt.Errorf(internal.TextColor.Red("Parameter [%s] not found after adding"), testCase.ParamKey)
			} else {
				if slices.Equal(retrievedValues, testCase.ParamValues) {
					tt.Logf("Special character parameter [%s] correctly preserved values", testCase.ParamKey)
				} else {
					tt.Errorf(internal.TextColor.Red("Special character parameter [%s] values changed. Expected: %#v, Got: %#v"), testCase.ParamKey, testCase.ParamValues, retrievedValues)
				}
			}
		})
	}
}

// Test case to validate Params key normalization and case sensitivity.
func Test_Params_KeyNormalization(t *testing.T) {
	testCases := []struct {
		Name       string
		AddKey     string
		GetKey     string
		Values     []string
		ShouldFind bool
	}{
		{"Exact key match", "name", "name", []string{"value"}, true},
		{"Key with leading/trailing spaces", "  name  ", "name", []string{"value"}, true},
		{"Different case keys", "Name", "name", []string{"value"}, false}, // Keys are case sensitive
		{"Key with internal spaces", "first name", "first name", []string{"value"}, true},
		{"Key with special characters", "user-id", "user-id", []string{"123"}, true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			params := make(internal.Params)
			params.Add(testCase.AddKey, testCase.Values)

			retrievedValues, ok := params.Get(testCase.GetKey)

			if ok == testCase.ShouldFind {
				if testCase.ShouldFind {
					if slices.Equal(retrievedValues, testCase.Values) {
						tt.Logf("Key normalization test passed for [%s] -> [%s]", testCase.AddKey, testCase.GetKey)
					} else {
						tt.Errorf(internal.TextColor.Red("Values don't match for key [%s]"), testCase.GetKey)
					}
				} else {
					tt.Logf("Key [%s] correctly not found when searching for [%s]", testCase.AddKey, testCase.GetKey)
				}
			} else {
				tt.Errorf(internal.TextColor.Red("Key normalization failed for [%s] -> [%s]. Expected found: %t, Got found: %t"), testCase.AddKey, testCase.GetKey, testCase.ShouldFind, ok)
			}
		})
	}
}

// Test case to validate Params.Add behavior with multiple calls for same key.
func Test_Params_MultipleAdds(t *testing.T) {
	params := make(internal.Params)

	// Add values to same key multiple times
	params.Add("tags", []string{"go", "web"})
	params.Add("tags", []string{"backend"})
	params.Add("tags", []string{"api", "rest"})

	// Add different key for comparison
	params.Add("category", []string{"programming"})

	testCases := []struct {
		Name               string
		ParamKey           string
		ExpectedValueCount int
		ExpectedValues     []string
	}{
		{"Multiple adds accumulate values", "tags", 5, []string{"go", "web", "backend", "api", "rest"}},
		{"Single add preserves single value", "category", 1, []string{"programming"}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			values, ok := params.Get(testCase.ParamKey)
			if !ok {
				tt.Errorf(internal.TextColor.Red("Parameter [%s] not found"), testCase.ParamKey)
				return
			}

			if len(values) == testCase.ExpectedValueCount {
				tt.Logf("Parameter [%s] has correct value count: %d", testCase.ParamKey, len(values))
			} else {
				tt.Errorf(internal.TextColor.Red("Parameter [%s] expected %d values, got %d"), testCase.ParamKey, testCase.ExpectedValueCount, len(values))
			}

			// Check if all expected values are present (order might vary)
			for _, expectedValue := range testCase.ExpectedValues {
				if !slices.Contains(values, expectedValue) {
					tt.Errorf(internal.TextColor.Red("Expected value [%s] not found in parameter [%s]"), expectedValue, testCase.ParamKey)
				}
			}

			tt.Logf("Parameter [%s] values: %#v", testCase.ParamKey, values)
		})
	}
}

// Test case to validate Params.Length() with various scenarios.
func Test_Params_Length_EdgeCases(t *testing.T) {
	params := make(internal.Params)

	testCases := []struct {
		Name           string
		Action         string
		ParamKey       string
		ParamValues    []string
		ExpectedLength int
	}{
		{"Initial length is zero", "check", "", nil, 0},
		{"Add first parameter", "add", "first", []string{"value1"}, 1},
		{"Add second parameter", "add", "second", []string{"value2"}, 2},
		{"Add to existing parameter", "add", "first", []string{"value3"}, 2}, // Length shouldn't change
		{"Add parameter with empty values", "add", "empty", []string{}, 3},
		{"Add parameter with nil-like behavior", "add", "another", []string{""}, 4},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			if testCase.Action == "add" {
				params.Add(testCase.ParamKey, testCase.ParamValues)
			}

			currentLength := params.Length()
			if currentLength == testCase.ExpectedLength {
				tt.Logf("Parameter length correctly: %d", currentLength)
			} else {
				tt.Errorf(internal.TextColor.Red("Expected length %d, got %d"), testCase.ExpectedLength, currentLength)
			}
		})
	}
}
