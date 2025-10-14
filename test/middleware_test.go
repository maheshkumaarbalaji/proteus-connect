package test

import (
	"testing"

	"github.com/citadelofcode/proteus/internal"
)

// Test case to validate the working of middleware processing logic.
func Test_MiddlewareProcessing(t *testing.T) {
	var MwWithStop = func(request *internal.HttpRequest, response *internal.HttpResponse, stop internal.StopFunction) {
		request.Server.Log("MwWithStop has been invoked", internal.INFO_LEVEL)
		stop()
	}
	var MwWithoutStop = func(request *internal.HttpRequest, response *internal.HttpResponse, stop internal.StopFunction) {
		request.Server.Log("MwWithoutStop has been invoked", internal.INFO_LEVEL)
	}

	testCases := []struct {
		Name           string
		MW             internal.Middleware
		ExpProcessNext bool
	}{
		{"Processing middleware with Stop Invocation", MwWithStop, false},
		{"Processing middleware without Stop Invocation", MwWithoutStop, true},
	}

	testServer := NewTestServer(t)
	testRequest := NewTestRequest(t, testServer, nil)
	testResponse := NewTestResponse(t, testRequest.Version, testServer, nil)

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(tt *testing.T) {
			mds := internal.CreateMiddlewares(testCase.MW)
			processNext := true
			for _, mw := range mds.Stack {
				mw(testRequest, testResponse, mds.Stop)
				processNext = mds.ProcessNext
				if !processNext {
					break
				}
			}
			if testCase.ExpProcessNext == processNext {
				tt.Log("The expected value for process next flag matches the value received after middleware processing")
			} else {
				tt.Error(internal.TextColor.Red("The expected value for process next flag does not match the value received after middleware processing"))
			}
		})
	}
}
