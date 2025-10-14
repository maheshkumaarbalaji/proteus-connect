package test

import (
	"testing"
	"time"

	"github.com/citadelofcode/proteus/internal"
)

// Test case to validate that unknown status codes return empty error content.
func Test_Status_GetErrorContent_Unknown(t *testing.T) {
	content := internal.StatusCode(799).GetErrorContent()
	if content != "" {
		t.Errorf(internal.TextColor.Red("Expected empty error content for unknown status, got %q"), content)
	}
}

// Test case to validate CLF time formatting helper.
func Test_GetCLFTime_Format(t *testing.T) {
	clfTime := internal.GetCLFTime()
	if clfTime == "" {
		t.Fatal(internal.TextColor.Red("Expected CLF time string, received empty value"))
	}
	if _, err := time.Parse("[02/Jan/2006:15:04:05 -0700]", clfTime); err != nil {
		t.Errorf(internal.TextColor.Red("Expected CLF time to match format, parse error: %v"), err)
	}
}

// Test case to validate IsMethodAllowed returns false for unsupported versions.
func Test_IsMethodAllowed_UnknownVersion(t *testing.T) {
	if internal.IsMethodAllowed("3.0", "GET") {
		t.Error(internal.TextColor.Red("Expected method to be disallowed for unknown HTTP version"))
	}
}
