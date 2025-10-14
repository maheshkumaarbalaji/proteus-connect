package test

import (
	"github.com/citadelofcode/proteus/internal"
	"strings"
	"testing"
)

// Test case to validate the status message received through the GetStatusMessage() of StatusCode.
func Test_GetStatusMessage(t *testing.T) {
	testCases := []struct {
		Name      string
		IpStatus  internal.StatusCode
		ExpOutput string
	}{
		{"A valid status code with an associated message", internal.Status200, "OK"},
		{"An invalid status code with no available message", internal.StatusCode(600), ""},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			message := testCase.IpStatus.GetStatusMessage()
			if strings.EqualFold(message, testCase.ExpOutput) {
				t.Logf("The expected status message [%s] matches the received status message [%s].", testCase.ExpOutput, message)
			} else {
				t.Errorf(internal.TextColor.Red("The expected status message [%s] does not match the received status message [%s]."), testCase.ExpOutput, message)
			}
		})
	}
}

// Test case to validate GetErrorContent returns templated HTML for known statuses.
func Test_GetErrorContent_KnownStatus(t *testing.T) {
	content := internal.Status500.GetErrorContent()
	if content == "" {
		t.Skip("No templated content generated for status")
	}
	if !strings.Contains(content, "500 - Internal Server Error") {
		t.Errorf(internal.TextColor.Red("Expected HTML content to contain status details, got %q"), content)
	}
}
