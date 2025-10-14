package test

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"reflect"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
	"unsafe"

	"github.com/citadelofcode/proteus/internal"
)

//go:linkname httpServerClose github.com/citadelofcode/proteus/internal.(*HttpServer).close
func httpServerClose(*internal.HttpServer)

//go:linkname httpServerIsClosed github.com/citadelofcode/proteus/internal.(*HttpServer).isClosed
func httpServerIsClosed(*internal.HttpServer) bool

//go:linkname httpServerProcessMiddlewares github.com/citadelofcode/proteus/internal.(*HttpServer).processMiddlewares
func httpServerProcessMiddlewares(*internal.HttpServer, *internal.HttpRequest, *internal.HttpResponse, []internal.Middleware) bool

//go:linkname httpServerAcceptConnections github.com/citadelofcode/proteus/internal.(*HttpServer).acceptConnections
func httpServerAcceptConnections(*internal.HttpServer)

//go:linkname httpServerHandleClient github.com/citadelofcode/proteus/internal.(*HttpServer).handleClient
func httpServerHandleClient(*internal.HttpServer, net.Conn)

//go:linkname httpServerGetKeepAliveHeuristic github.com/citadelofcode/proteus/internal.(*HttpServer).getKeepAliveHeuristic
func httpServerGetKeepAliveHeuristic(*internal.HttpServer, int) (int, int)

//go:linkname httpServerTerminate github.com/citadelofcode/proteus/internal.(*HttpServer).terminate
func httpServerTerminate(*internal.HttpServer)

//go:linkname httpServerLogStatus github.com/citadelofcode/proteus/internal.(*HttpServer).logStatus
func httpServerLogStatus(*internal.HttpServer, *internal.HttpRequest, *internal.HttpResponse)

type stubAddr struct{}

func (stubAddr) Network() string { return "stub" }
func (stubAddr) String() string  { return "stub-addr" }

type stubListener struct {
	mu        sync.Mutex
	closed    bool
	acceptErr error
}

func (s *stubListener) Accept() (net.Conn, error) {
	time.Sleep(5 * time.Millisecond)
	return nil, s.acceptErr
}

func (s *stubListener) Close() error {
	s.mu.Lock()
	s.closed = true
	s.mu.Unlock()
	return nil
}

func (s *stubListener) Addr() net.Addr {
	return stubAddr{}
}

type queueListener struct {
	mu     sync.Mutex
	closed bool
	connCh chan net.Conn
}

func (q *queueListener) Accept() (net.Conn, error) {
	conn, ok := <-q.connCh
	if !ok {
		return nil, io.EOF
	}
	return conn, nil
}

func (q *queueListener) Close() error {
	q.mu.Lock()
	q.closed = true
	q.mu.Unlock()
	return nil
}

func (q *queueListener) Addr() net.Addr {
	return stubAddr{}
}

func setServerField(t *testing.T, server *internal.HttpServer, field string, value any) {
	t.Helper()
	rv := reflect.ValueOf(server).Elem()
	fv := rv.FieldByName(field)
	if !fv.IsValid() {
		t.Fatalf(internal.TextColor.Red("Field %s does not exist on HttpServer"), field)
	}
	reflect.NewAt(fv.Type(), unsafe.Pointer(fv.UnsafeAddr())).Elem().Set(reflect.ValueOf(value))
}

func getServerListener(server *internal.HttpServer) net.Listener {
	rv := reflect.ValueOf(server).Elem().FieldByName("listener")
	if !rv.IsValid() || rv.IsNil() {
		return nil
	}
	return reflect.NewAt(rv.Type(), unsafe.Pointer(rv.UnsafeAddr())).Elem().Interface().(net.Listener)
}

// Test case to validate close/isClosed behaviour on HttpServer.
func Test_HttpServer_CloseAndIsClosed(t *testing.T) {
	server := internal.NewServer("", 0)
	listener := &stubListener{}
	setServerField(t, server, "listener", net.Listener(listener))

	httpServerClose(server)
	if !listener.closed {
		t.Error(internal.TextColor.Red("Expected listener.Close to be invoked"))
	}
	if !httpServerIsClosed(server) {
		t.Error(internal.TextColor.Red("Expected server to be marked closed"))
	}
}

// Test case to validate server middleware execution path.
func Test_HttpServer_ProcessMiddlewares_Internal(t *testing.T) {
	server := internal.NewServer("", 0)
	request := NewTestRequest(t, server, nil)
	response := NewTestResponse(t, "1.1", server, nil)

	var callOrder []string
	middlewares := []internal.Middleware{
		func(req *internal.HttpRequest, res *internal.HttpResponse, stop internal.StopFunction) {
			callOrder = append(callOrder, "first")
		},
		func(req *internal.HttpRequest, res *internal.HttpResponse, stop internal.StopFunction) {
			callOrder = append(callOrder, "second")
			stop()
		},
		func(req *internal.HttpRequest, res *internal.HttpResponse, stop internal.StopFunction) {
			callOrder = append(callOrder, "third")
		},
	}

	halted := httpServerProcessMiddlewares(server, request, response, middlewares)
	if !halted {
		t.Error(internal.TextColor.Red("Expected middleware processing to halt when stop was invoked"))
	}

	if len(callOrder) != 2 {
		t.Errorf(internal.TextColor.Red("Expected two middlewares to execute, got %v"), callOrder)
	}
}

// Test case to validate keep-alive heuristic outputs.
func Test_HttpServer_GetKeepAliveHeuristic_Internal(t *testing.T) {
	server := internal.NewServer("", 0)
	timeout, max := httpServerGetKeepAliveHeuristic(server, 1)
	if timeout <= 0 || max != 100 {
		t.Errorf(internal.TextColor.Red("Unexpected heuristic values: timeout=%d max=%d"), timeout, max)
	}

	timeoutHigh, _ := httpServerGetKeepAliveHeuristic(server, 50)
	if timeoutHigh > timeout {
		t.Errorf(internal.TextColor.Red("Expected timeout to decrease with higher connection count, got %d > %d"), timeoutHigh, timeout)
	}
}

// Test case to validate server termination logic.
func Test_HttpServer_Terminate_Internal(t *testing.T) {
	server := internal.NewServer("", 0)
	serverLogger := log.New(&bytes.Buffer{}, "", 0)
	setServerField(t, server, "requestLogger", serverLogger)

	listener := &stubListener{}
	setServerField(t, server, "listener", net.Listener(listener))

	shutdown := make(chan struct{})
	setServerField(t, server, "shutdown", shutdown)

	setServerField(t, server, "cw", new(internal.ConnectionWatcher))

	httpServerTerminate(server)
	if !listener.closed {
		t.Error(internal.TextColor.Red("Expected terminate to close listener"))
	}
	if !httpServerIsClosed(server) {
		t.Error(internal.TextColor.Red("Expected terminate to mark server closed"))
	}
}

// Test case to validate server termination timeout branch.
func Test_HttpServer_Terminate_Timeout(t *testing.T) {
	server := internal.NewServer("", 0)
	setServerField(t, server, "requestLogger", log.New(&bytes.Buffer{}, "", 0))

	shutdown := make(chan struct{})
	setServerField(t, server, "shutdown", shutdown)

	listener := &stubListener{}
	setServerField(t, server, "listener", net.Listener(listener))

	setServerField(t, server, "cw", new(internal.ConnectionWatcher))

	originalTimeout := internal.ServerDefaults["shutdown_timeout"].(int)
	internal.ServerDefaults["shutdown_timeout"] = 0
	defer func() { internal.ServerDefaults["shutdown_timeout"] = originalTimeout }()

	wgValue := reflect.ValueOf(server).Elem().FieldByName("wg")
	wg := (*sync.WaitGroup)(unsafe.Pointer(wgValue.UnsafeAddr()))
	wg.Add(1) // Intentionally never done to trigger timeout path.

	httpServerTerminate(server)
}

// Test case to validate logStatus across supported formats.
func Test_HttpServer_LogStatus_Internal(t *testing.T) {
	server := internal.NewServer("", 0)
	buffer := &bytes.Buffer{}
	setServerField(t, server, "requestLogger", log.New(buffer, "", 0))

	request := NewTestRequest(t, server, nil)
	request.Method = "GET"
	request.ResourcePath = "/status"
	request.Version = "1.1"
	request.ClientAddress = "127.0.0.1"
	request.Locals["Started"] = time.Now().Add(-25 * time.Millisecond)

	response := NewTestResponse(t, "1.1", server, bufio.NewWriter(&bytes.Buffer{}))

	formats := []struct {
		format     string
		statusCode int
	}{
		{internal.DEV_LOGGER, 200},
		{internal.TINY_LOGGER, 302},
		{internal.SHORT_LOGGER, 404},
		{internal.COMMON_LOGGER, 503},
		{"", 101},
	}

	for _, testCase := range formats {
		setServerField(t, server, "logFormat", testCase.format)
		response.StatusCode = testCase.statusCode
		httpServerLogStatus(server, request, response)
	}

	logOutput := buffer.String()
	for _, keyword := range []string{"GET", "/status", "HTTP/1.1"} {
		if !strings.Contains(logOutput, keyword) {
			t.Errorf(internal.TextColor.Red("Expected log output to contain %s"), keyword)
		}
	}
}

// Test case to validate acceptConnections reacts to shutdown signal.
func Test_HttpServer_AcceptConnections_Internal(t *testing.T) {
	server := internal.NewServer("", 0)
	setServerField(t, server, "requestLogger", log.New(&bytes.Buffer{}, "", 0))
	setServerField(t, server, "listener", net.Listener(&stubListener{acceptErr: errors.New("accept failure")}))
	setServerField(t, server, "cw", new(internal.ConnectionWatcher))

	shutdown := make(chan struct{})
	setServerField(t, server, "shutdown", shutdown)

	wgValue := reflect.ValueOf(server).Elem().FieldByName("wg")
	wg := (*sync.WaitGroup)(unsafe.Pointer(wgValue.UnsafeAddr()))
	wg.Add(1)

	done := make(chan struct{})
	go func() {
		httpServerAcceptConnections(server)
		close(done)
	}()

	time.Sleep(15 * time.Millisecond)
	close(shutdown)
	wg.Wait()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal(internal.TextColor.Red("acceptConnections did not exit after shutdown"))
	}
}

// Test case to validate handleClient end-to-end request processing.
func Test_HttpServer_HandleClient_Internal(t *testing.T) {
	server := internal.NewServer("", 0)
	setServerField(t, server, "requestLogger", log.New(&bytes.Buffer{}, "", 0))
	setServerField(t, server, "cw", new(internal.ConnectionWatcher))

	server.Use(func(req *internal.HttpRequest, res *internal.HttpResponse, stop internal.StopFunction) {
		req.Locals["server-middleware"] = true
	})

	err := server.Router.Get("/hello", func(req *internal.HttpRequest, res *internal.HttpResponse) {
		if req.Locals["server-middleware"] != true {
			t.Error(internal.TextColor.Red("Expected server middleware to set locals"))
		}
		res.Status(internal.Status200)
		res.Send("handled")
	}, func(req *internal.HttpRequest, res *internal.HttpResponse, stop internal.StopFunction) {
		req.Locals["route-middleware"] = true
	})
	if err != nil {
		t.Fatalf(internal.TextColor.Red("Unexpected error registering route: %v"), err)
	}

	serverConn, clientConn := net.Pipe()
	defer clientConn.Close()

	wgValue := reflect.ValueOf(server).Elem().FieldByName("wg")
	wg := (*sync.WaitGroup)(unsafe.Pointer(wgValue.UnsafeAddr()))
	wg.Add(1)

	go httpServerHandleClient(server, serverConn)

	requestData := "GET /hello HTTP/1.1\r\nHost: example.com\r\nConnection: close\r\n\r\n"
	if _, err := clientConn.Write([]byte(requestData)); err != nil {
		t.Fatalf(internal.TextColor.Red("Failed to write request: %v"), err)
	}

	buffer := make([]byte, 256)
	clientConn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	n, err := clientConn.Read(buffer)
	if err != nil && !errors.Is(err, net.ErrClosed) && !errors.Is(err, io.EOF) {
		t.Fatalf(internal.TextColor.Red("Failed to read response: %v"), err)
	}

	responseText := string(buffer[:n])
	if !strings.Contains(responseText, "HTTP/1.1 200 OK") {
		t.Errorf(internal.TextColor.Red("Expected HTTP 200 response, got %q"), responseText)
	}

	clientConn.Close()
	wg.Wait()
}

// Test case to validate keep-alive processing with multiple requests.
func Test_HttpServer_HandleClient_KeepAlive(t *testing.T) {
	server := internal.NewServer("", 0)
	setServerField(t, server, "requestLogger", log.New(&bytes.Buffer{}, "", 0))
	setServerField(t, server, "cw", new(internal.ConnectionWatcher))

	server.Use(func(req *internal.HttpRequest, res *internal.HttpResponse, stop internal.StopFunction) {
		req.Locals["hit"] = true
	})

	err := server.Router.Get("/hello", func(req *internal.HttpRequest, res *internal.HttpResponse) {
		if req.Locals["hit"] != true {
			t.Error(internal.TextColor.Red("Expected server middleware to execute for keep-alive request"))
		}
		res.Status(internal.Status200)
		res.Send("keep-alive")
	})
	if err != nil {
		t.Fatalf(internal.TextColor.Red("Unexpected error registering route: %v"), err)
	}

	serverConn, clientConn := net.Pipe()
	defer clientConn.Close()

	wgValue := reflect.ValueOf(server).Elem().FieldByName("wg")
	wg := (*sync.WaitGroup)(unsafe.Pointer(wgValue.UnsafeAddr()))
	wg.Add(1)

	go httpServerHandleClient(server, serverConn)

	firstRequest := "GET /hello HTTP/1.1\r\nHost: example.com\r\nConnection: keep-alive\r\n\r\n"
	if _, err := clientConn.Write([]byte(firstRequest)); err != nil {
		t.Fatalf(internal.TextColor.Red("Failed to write keep-alive request: %v"), err)
	}

	clientConn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	buf := make([]byte, 512)
	n, err := clientConn.Read(buf)
	if err != nil {
		t.Fatalf(internal.TextColor.Red("Failed to read keep-alive response: %v"), err)
	}
	firstResponse := string(buf[:n])
	if !strings.Contains(firstResponse, "Connection: keep-alive") {
		t.Errorf(internal.TextColor.Red("Expected keep-alive headers in response, got %q"), firstResponse)
	}

	secondRequest := "GET /hello HTTP/1.1\r\nHost: example.com\r\nConnection: close\r\n\r\n"
	if _, err := clientConn.Write([]byte(secondRequest)); err != nil {
		t.Fatalf(internal.TextColor.Red("Failed to write closing request: %v"), err)
	}

	clientConn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	n, err = clientConn.Read(buf)
	if err != nil && !errors.Is(err, net.ErrClosed) && !errors.Is(err, io.EOF) {
		t.Fatalf(internal.TextColor.Red("Failed to read second response: %v"), err)
	}
	secondResponse := string(buf[:n])
	if !strings.Contains(secondResponse, "keep-alive") {
		t.Errorf(internal.TextColor.Red("Expected response body for second request, got %q"), secondResponse)
	}

	clientConn.Close()
	wg.Wait()
}

// Test case to validate keep-alive logic with real TCP connection to trigger socket configuration.
func Test_HttpServer_HandleClient_KeepAliveTCP(t *testing.T) {
	server := internal.NewServer("", 0)
	setServerField(t, server, "requestLogger", log.New(&bytes.Buffer{}, "", 0))
	setServerField(t, server, "cw", new(internal.ConnectionWatcher))

	server.Router.Get("/hello", func(req *internal.HttpRequest, res *internal.HttpResponse) {
		res.Status(internal.Status200)
		res.Send("tcp")
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skip("Unable to create TCP listener")
	}
	defer ln.Close()

	clientConn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf(internal.TextColor.Red("Failed to dial TCP listener: %v"), err)
	}

	serverConn, err := ln.Accept()
	if err != nil {
		t.Fatalf(internal.TextColor.Red("Failed to accept TCP connection: %v"), err)
	}

	wgValue := reflect.ValueOf(server).Elem().FieldByName("wg")
	wg := (*sync.WaitGroup)(unsafe.Pointer(wgValue.UnsafeAddr()))
	wg.Add(1)

	go httpServerHandleClient(server, serverConn)

	request := "GET /hello HTTP/1.1\r\nHost: example.com\r\nConnection: keep-alive\r\n\r\n"
	if _, err := clientConn.Write([]byte(request)); err != nil {
		t.Fatalf(internal.TextColor.Red("Failed to write keep-alive request: %v"), err)
	}

	buffer := make([]byte, 256)
	clientConn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	if _, err := clientConn.Read(buffer); err != nil {
		t.Fatalf(internal.TextColor.Red("Failed to read response: %v"), err)
	}

	clientConn.Close()
	wg.Wait()
}

// Test case to validate method not allowed handling inside handleClient.
func Test_HttpServer_HandleClient_MethodNotAllowed(t *testing.T) {
	server := internal.NewServer("", 0)
	setServerField(t, server, "requestLogger", log.New(&bytes.Buffer{}, "", 0))
	setServerField(t, server, "cw", new(internal.ConnectionWatcher))

	serverConn, clientConn := net.Pipe()
	defer clientConn.Close()

	wgValue := reflect.ValueOf(server).Elem().FieldByName("wg")
	wg := (*sync.WaitGroup)(unsafe.Pointer(wgValue.UnsafeAddr()))
	wg.Add(1)

	go httpServerHandleClient(server, serverConn)

	request := "PURGE /hello HTTP/1.1\r\nHost: example.com\r\nConnection: close\r\n\r\n"
	if _, err := clientConn.Write([]byte(request)); err != nil {
		t.Fatalf(internal.TextColor.Red("Failed to write PURGE request: %v"), err)
	}

	clientConn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	buffer := make([]byte, 512)
	n, err := clientConn.Read(buffer)
	if err != nil && !errors.Is(err, net.ErrClosed) && !errors.Is(err, io.EOF) {
		t.Fatalf(internal.TextColor.Red("Failed to read PURGE response: %v"), err)
	}

	response := string(buffer[:n])
	if !strings.Contains(response, "405 Method Not Allowed") {
		t.Errorf(internal.TextColor.Red("Expected 405 response, got %q"), response)
	}

	clientConn.Close()
	wg.Wait()
}

// Test case to validate 404 handling when route is not found.
func Test_HttpServer_HandleClient_RouteNotFound(t *testing.T) {
	server := internal.NewServer("", 0)
	setServerField(t, server, "requestLogger", log.New(&bytes.Buffer{}, "", 0))
	setServerField(t, server, "cw", new(internal.ConnectionWatcher))

	serverConn, clientConn := net.Pipe()
	defer clientConn.Close()

	wgValue := reflect.ValueOf(server).Elem().FieldByName("wg")
	wg := (*sync.WaitGroup)(unsafe.Pointer(wgValue.UnsafeAddr()))
	wg.Add(1)

	go httpServerHandleClient(server, serverConn)

	request := "GET /missing HTTP/1.1\r\nHost: example.com\r\nConnection: close\r\n\r\n"
	if _, err := clientConn.Write([]byte(request)); err != nil {
		t.Fatalf(internal.TextColor.Red("Failed to write missing route request: %v"), err)
	}

	clientConn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	buffer := make([]byte, 512)
	n, err := clientConn.Read(buffer)
	if err != nil && !errors.Is(err, net.ErrClosed) && !errors.Is(err, io.EOF) {
		t.Fatalf(internal.TextColor.Red("Failed to read missing route response: %v"), err)
	}

	response := string(buffer[:n])
	if !strings.Contains(response, "404 Not Found") {
		t.Errorf(internal.TextColor.Red("Expected 404 response, got %q"), response)
	}

	clientConn.Close()
	wg.Wait()
}

// Test case to validate acceptConnections handling a real client before shutdown.
func Test_HttpServer_AcceptConnections_WithClient(t *testing.T) {
	server := internal.NewServer("", 0)
	setServerField(t, server, "requestLogger", log.New(&bytes.Buffer{}, "", 0))
	setServerField(t, server, "cw", new(internal.ConnectionWatcher))

	listener := &queueListener{connCh: make(chan net.Conn, 1)}
	setServerField(t, server, "listener", net.Listener(listener))

	shutdown := make(chan struct{})
	setServerField(t, server, "shutdown", shutdown)

	err := server.Router.Get("/hello", func(req *internal.HttpRequest, res *internal.HttpResponse) {
		res.Status(internal.Status200)
		res.Send("ok")
	})
	if err != nil {
		t.Fatalf(internal.TextColor.Red("Unexpected error registering route: %v"), err)
	}

	wgValue := reflect.ValueOf(server).Elem().FieldByName("wg")
	wg := (*sync.WaitGroup)(unsafe.Pointer(wgValue.UnsafeAddr()))
	wg.Add(1)

	go httpServerAcceptConnections(server)

	serverConn, clientConn := net.Pipe()
	defer clientConn.Close()

	listener.connCh <- serverConn

	request := "GET /hello HTTP/1.1\r\nHost: example.com\r\nConnection: close\r\n\r\n"
	if _, err := clientConn.Write([]byte(request)); err != nil {
		t.Fatalf(internal.TextColor.Red("Failed to write request: %v"), err)
	}

	clientConn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	buffer := make([]byte, 256)
	n, err := clientConn.Read(buffer)
	if err != nil && !errors.Is(err, net.ErrClosed) && !errors.Is(err, io.EOF) {
		t.Fatalf(internal.TextColor.Red("Failed to read response: %v"), err)
	}

	if !strings.Contains(string(buffer[:n]), "200 OK") {
		t.Errorf(internal.TextColor.Red("Expected 200 response, got %q"), string(buffer[:n]))
	}

	close(listener.connCh)
	close(shutdown)
	wg.Wait()
}

// Test case to validate route middleware short-circuiting request handling.
func Test_HttpServer_HandleClient_RouteMiddlewareStop(t *testing.T) {
	server := internal.NewServer("", 0)
	setServerField(t, server, "requestLogger", log.New(&bytes.Buffer{}, "", 0))
	setServerField(t, server, "cw", new(internal.ConnectionWatcher))

	handlerInvoked := false
	err := server.Router.Get("/mw", func(req *internal.HttpRequest, res *internal.HttpResponse) {
		handlerInvoked = true
	}, func(req *internal.HttpRequest, res *internal.HttpResponse, stop internal.StopFunction) {
		res.Status(internal.Status202)
		_ = res.Send("middleware")
		stop()
	}, func(req *internal.HttpRequest, res *internal.HttpResponse, stop internal.StopFunction) {
		t.Error(internal.TextColor.Red("Second middleware should not execute after stop"))
	})
	if err != nil {
		t.Fatalf(internal.TextColor.Red("Unexpected error registering route: %v"), err)
	}

	serverConn, clientConn := net.Pipe()
	defer clientConn.Close()

	wgValue := reflect.ValueOf(server).Elem().FieldByName("wg")
	wg := (*sync.WaitGroup)(unsafe.Pointer(wgValue.UnsafeAddr()))
	wg.Add(1)

	go httpServerHandleClient(server, serverConn)

	request := "GET /mw HTTP/1.1\r\nHost: example.com\r\nConnection: close\r\n\r\n"
	if _, err := clientConn.Write([]byte(request)); err != nil {
		t.Fatalf(internal.TextColor.Red("Failed to write request: %v"), err)
	}

	clientConn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	buffer := make([]byte, 512)
	n, err := clientConn.Read(buffer)
	if err != nil && !errors.Is(err, net.ErrClosed) && !errors.Is(err, io.EOF) {
		t.Fatalf(internal.TextColor.Red("Failed to read middleware response: %v"), err)
	}

	response := string(buffer[:n])
	if !strings.Contains(response, "202 Accepted") || !strings.Contains(response, "middleware") {
		t.Errorf(internal.TextColor.Red("Expected middleware-generated response, got %q"), response)
	}
	if handlerInvoked {
		t.Error(internal.TextColor.Red("Route handler should not execute when middleware stops processing"))
	}

	clientConn.Close()
	wg.Wait()
}

// Test case to ensure malformed requests are handled gracefully.
func Test_HttpServer_HandleClient_InvalidRequest(t *testing.T) {
	server := internal.NewServer("", 0)
	setServerField(t, server, "requestLogger", log.New(&bytes.Buffer{}, "", 0))
	setServerField(t, server, "cw", new(internal.ConnectionWatcher))

	serverConn, clientConn := net.Pipe()
	defer clientConn.Close()

	wgValue := reflect.ValueOf(server).Elem().FieldByName("wg")
	wg := (*sync.WaitGroup)(unsafe.Pointer(wgValue.UnsafeAddr()))
	wg.Add(1)

	go httpServerHandleClient(server, serverConn)

	if _, err := clientConn.Write([]byte("BROKEN REQUEST")); err != nil {
		t.Fatalf(internal.TextColor.Red("Failed to write invalid request: %v"), err)
	}

	clientConn.Close()
	wg.Wait()
}

// Test case to validate Listen error handling for invalid addresses.
func Test_HttpServer_Listen_InvalidAddress(t *testing.T) {
	server := internal.NewServer("256.256.256.256", 8080)
	buffer := &bytes.Buffer{}
	setServerField(t, server, "requestLogger", log.New(buffer, "", 0))

	server.Listen()

	if !strings.Contains(buffer.String(), "Error occurred while setting up listener socket") {
		t.Errorf(internal.TextColor.Red("Expected listen error to be logged, got %q"), buffer.String())
	}
}

// Test case to cover successful listen lifecycle with a real HTTP request before shutdown.
func Test_HttpServer_Listen_E2ERequest(t *testing.T) {
	server := internal.NewServer("127.0.0.1", 0)
	server.PortNumber = 0
	buffer := &bytes.Buffer{}
	setServerField(t, server, "requestLogger", log.New(buffer, "", 0))

	err := server.Router.Get("/hello", func(req *internal.HttpRequest, res *internal.HttpResponse) {
		res.Status(internal.Status200)
		res.Send("hello")
	})
	if err != nil {
		t.Fatalf(internal.TextColor.Red("Unexpected error registering route: %v"), err)
	}

	done := make(chan struct{})
	go func() {
		server.Listen()
		close(done)
	}()

	var addr string
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if listener := getServerListener(server); listener != nil {
			addr = listener.Addr().String()
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if addr == "" {
		t.Skip("Listener did not start in time")
	}

	conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
	if err != nil {
		t.Fatalf(internal.TextColor.Red("Failed to dial server: %v"), err)
	}

	request := "GET /hello HTTP/1.1\r\nHost: example.com\r\nConnection: close\r\n\r\n"
	if _, err := conn.Write([]byte(request)); err != nil {
		t.Fatalf(internal.TextColor.Red("Failed to write request: %v"), err)
	}

	conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	responseBuf := make([]byte, 256)
	if _, err := conn.Read(responseBuf); err != nil {
		t.Fatalf(internal.TextColor.Red("Failed to read response: %v"), err)
	}
	conn.Close()

	if err := syscall.Kill(os.Getpid(), syscall.SIGINT); err != nil {
		t.Fatalf(internal.TextColor.Red("Failed to send SIGINT: %v"), err)
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal(internal.TextColor.Red("Server did not shut down after signal"))
	}
}
