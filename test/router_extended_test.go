package test

import (
	"testing"

	"github.com/citadelofcode/proteus/internal"
)

// Test case to validate the additional HTTP verb helpers on Router.
func Test_Router_AdditionalVerbs(t *testing.T) {
	router := NewTestRouter(t)
	server := NewTestServer(t)

	var invocations []string
	handler := func(name string) internal.RouteHandler {
		return func(request *internal.HttpRequest, response *internal.HttpResponse) {
			invocations = append(invocations, name)
		}
	}

	middleware := func(name string) internal.Middleware {
		return func(request *internal.HttpRequest, response *internal.HttpResponse, stop internal.StopFunction) {
			invocations = append(invocations, name)
		}
	}

	if err := router.Head("/status", handler("HEAD"), middleware("HEAD-mw")); err != nil {
		t.Fatalf(internal.TextColor.Red("Unexpected error registering HEAD route: %v"), err)
	}
	if err := router.Put("/resource/:id", handler("PUT")); err != nil {
		t.Fatalf(internal.TextColor.Red("Unexpected error registering PUT route: %v"), err)
	}
	if err := router.Trace("/trace", handler("TRACE")); err != nil {
		t.Fatalf(internal.TextColor.Red("Unexpected error registering TRACE route: %v"), err)
	}
	if err := router.Options("/options", handler("OPTIONS")); err != nil {
		t.Fatalf(internal.TextColor.Red("Unexpected error registering OPTIONS route: %v"), err)
	}
	if err := router.Connect("/tunnel", handler("CONNECT")); err != nil {
		t.Fatalf(internal.TextColor.Red("Unexpected error registering CONNECT route: %v"), err)
	}

	// Verify matching for HEAD route executes middleware and handler.
	request := NewTestRequest(t, server, nil)
	request.Method = "HEAD"
	request.ResourcePath = "/status"
	route, err := router.Match(request)
	if err != nil {
		t.Fatalf(internal.TextColor.Red("Did not expect match error: %v"), err)
	}

	for _, mw := range route.Middlewares {
		mw(request, NewTestResponse(t, "1.1", server, nil), func() {})
	}
	route.RouteHandler(request, NewTestResponse(t, "1.1", server, nil))

	if len(invocations) != 2 || invocations[0] != "HEAD-mw" || invocations[1] != "HEAD" {
		t.Errorf(internal.TextColor.Red("Expected middleware and handler invocation order, got %v"), invocations)
	}
}
