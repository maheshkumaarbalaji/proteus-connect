package test

import (
	"strings"
	"testing"
	"github.com/citadelofcode/proteus/internal"
)

// Test case to validate the working of all color functions in TextColor.
func Test_TextColor_AllColors(t *testing.T) {
	testCases := []struct {
		Name string
		InputText string
		ColorFunc func(string) string
		ExpectedPrefix string
		ExpectedSuffix string
	} {
		{ "Black color function", "test text", internal.TextColor.Black, "\033[30m", "\033[0m" },
		{ "Red color function", "error message", internal.TextColor.Red, "\033[31m", "\033[0m" },
		{ "Green color function", "success message", internal.TextColor.Green, "\033[32m", "\033[0m" },
		{ "Yellow color function", "warning message", internal.TextColor.Yellow, "\033[33m", "\033[0m" },
		{ "Blue color function", "info message", internal.TextColor.Blue, "\033[34m", "\033[0m" },
		{ "Magenta color function", "debug message", internal.TextColor.Magenta, "\033[35m", "\033[0m" },
		{ "Cyan color function", "highlight message", internal.TextColor.Cyan, "\033[36m", "\033[0m" },
		{ "White color function", "default message", internal.TextColor.White, "\033[37m", "\033[0m" },
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			result := testCase.ColorFunc(testCase.InputText)
			
			// Check if the result starts with the expected ANSI escape sequence
			if !strings.HasPrefix(result, testCase.ExpectedPrefix) {
				tt.Errorf(internal.TextColor.Red("Expected result to start with [%s], but got [%s]"), testCase.ExpectedPrefix, result[:len(testCase.ExpectedPrefix)])
			} else {
				tt.Logf("Result correctly starts with expected ANSI escape sequence [%s]", testCase.ExpectedPrefix)
			}
			
			// Check if the result ends with the reset sequence
			if !strings.HasSuffix(result, testCase.ExpectedSuffix) {
				tt.Errorf(internal.TextColor.Red("Expected result to end with [%s], but got [%s]"), testCase.ExpectedSuffix, result[len(result)-len(testCase.ExpectedSuffix):])
			} else {
				tt.Logf("Result correctly ends with expected reset sequence [%s]", testCase.ExpectedSuffix)
			}
			
			// Check if the original text is preserved in the middle
			expectedResult := testCase.ExpectedPrefix + testCase.InputText + testCase.ExpectedSuffix
			if result != expectedResult {
				tt.Errorf(internal.TextColor.Red("Expected complete result [%s], but got [%s]"), expectedResult, result)
			} else {
				tt.Logf("Complete colored text matches expected format: [%s]", result)
			}
		})
	}
}

// Test case to validate color functions with special characters and edge cases.
func Test_TextColor_EdgeCases(t *testing.T) {
	testCases := []struct {
		Name string
		InputText string
		ColorFunc func(string) string
	} {
		{ "Empty string input", "", internal.TextColor.Red },
		{ "String with spaces", "  hello world  ", internal.TextColor.Green },
		{ "String with newlines", "line1\nline2\nline3", internal.TextColor.Blue },
		{ "String with tabs", "col1\tcol2\tcol3", internal.TextColor.Yellow },
		{ "String with special characters", "!@#$%^&*()", internal.TextColor.Magenta },
		{ "String with unicode characters", "Hello 世界 🌍", internal.TextColor.Cyan },
		{ "Very long string", strings.Repeat("A", 1000), internal.TextColor.White },
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			result := testCase.ColorFunc(testCase.InputText)
			
			// Ensure the result is not empty (unless input was empty)
			if len(result) == 0 && len(testCase.InputText) > 0 {
				tt.Error(internal.TextColor.Red("Expected non-empty result for non-empty input"))
				return
			}
			
			// Ensure the original text is contained within the result
			if len(testCase.InputText) > 0 && !strings.Contains(result, testCase.InputText) {
				tt.Errorf(internal.TextColor.Red("Expected result to contain original text [%s], but got [%s]"), testCase.InputText, result)
			} else {
				tt.Logf("Original text [%s] is correctly preserved in colored result", testCase.InputText)
			}
		})
	}
}

// Test case to validate that different color functions produce different outputs for the same input.
func Test_TextColor_DifferentColorsProduceDifferentOutputs(t *testing.T) {
	testText := "sample text"
	
	colors := []struct {
		Name string
		Func func(string) string
	}{
		{"Black", internal.TextColor.Black},
		{"Red", internal.TextColor.Red},
		{"Green", internal.TextColor.Green},
		{"Yellow", internal.TextColor.Yellow},
		{"Blue", internal.TextColor.Blue},
		{"Magenta", internal.TextColor.Magenta},
		{"Cyan", internal.TextColor.Cyan},
		{"White", internal.TextColor.White},
	}
	
	results := make(map[string]string)
	
	// Generate results for all colors
	for _, color := range colors {
		results[color.Name] = color.Func(testText)
	}
	
	// Compare each color result with every other color result
	for i, color1 := range colors {
		for j, color2 := range colors {
			if i != j {
				result1 := results[color1.Name]
				result2 := results[color2.Name]
				
				if result1 == result2 {
					t.Errorf(internal.TextColor.Red("Color functions [%s] and [%s] produced identical results [%s], expected different outputs"), color1.Name, color2.Name, result1)
				}
			}
		}
	}
	
	t.Log("All color functions produce unique outputs as expected")
}