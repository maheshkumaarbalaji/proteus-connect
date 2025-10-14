package internal

import (
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"os"
	"os/signal"
	"runtime"
	"slices"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Number of CPUs that can be used for facilitating concurrency.
var numCPU = runtime.NumCPU()

// Structure to track all the active connections maintained by the server.
type ConnectionWatcher struct {
	// Mutex to synchronize read-write activities for tracking the number of active connections.
	mu sync.RWMutex
	// Contains the number of active connections with the server.
	connCount int
}

// UpdateCount increases the connection count by the specified delta.
// Used internally to track active connections in a thread-safe manner.
// Positive delta adds connections, negative delta removes connections.
func (cw *ConnectionWatcher) UpdateCount(delta int) {
	cw.mu.Lock()
	cw.connCount += delta
	cw.mu.Unlock()
}

// GetCount returns the current number of active connections.
// Thread-safe method providing read-only access to connection count.
// Returns the number of concurrent client connections to the server.
func (cw *ConnectionWatcher) GetCount() int {
	cw.mu.RLock()
	count := cw.connCount
	cw.mu.RUnlock()
	return count
}

// Structure to create an instance of a web server.
type HttpServer struct {
	// Hostname of the web server instance.
	HostAddress string
	// Port number where web server instance is listening for incoming requests.
	PortNumber int
	// // key-value pairs to hold variables available for all requests and responses processed by the server instance.
	Locals map[string]any
	// Listener created and bound to the host address and port number.
	listener net.Listener
	// Router instance that contains all the routes and their associated handlers.
	Router *Router
	// Logger to capture request processing logs.
	requestLogger *log.Logger
	// Format in which the logs will be captured.
	logFormat string
	// Waitgroup to synchronize the termination of all request connections.
	wg sync.WaitGroup
	// Channel to transmit server shutdown signal across all goroutines.
	shutdown chan struct{}
	// Instance of ConnectionWatcher.
	cw *ConnectionWatcher
	// Flag to determine if the listener is closed.
	listClosed bool
	// Mutex to manage read-write activities on the listClosed flag.
	limu sync.RWMutex
	// Server level middlewares to be executed for all incoming requests regardless of the matching route.
	middlewares []Middleware
}

// Function that closes the server listener and marks the listClosed flag as closed.
func (srv *HttpServer) close() {
	srv.limu.Lock()
	srv.listClosed = true
	srv.limu.Unlock()
	srv.listener.Close()
}

// Returns true if the server listener is already closed and false, otherwise.
func (srv *HttpServer) isClosed() bool {
	isClose := false
	srv.limu.RLock()
	isClose = srv.listClosed
	srv.limu.RUnlock()
	return isClose
}

// Processes the given set of middlewares for the request and response and returns a boolean value to confirm if the request processing has completed with a response sent back to the client.
func (srv *HttpServer) processMiddlewares(request *HttpRequest, response *HttpResponse, middlewareList []Middleware) bool {
	mwsInstance := CreateMiddlewares(middlewareList...)
	for _, middleware := range mwsInstance.Stack {
		if !mwsInstance.ProcessNext {
			return true
		}
		middleware(request, response, mwsInstance.Stop)
	}
	return false
}

// Accepts incoming connections and creates seperate goroutines for each new client.
func (srv *HttpServer) acceptConnections() {
	defer srv.wg.Done()

	for {
		select {
		case <-srv.shutdown:
			srv.Log("Server Shutdown initiated :: No new connections will be accepted from now.", WARN_LEVEL)
			return
		default:
			clientConnection, err := srv.listener.Accept()
			if err != nil {
				if !srv.isClosed() {
					srv.Log(fmt.Sprintf("Error occurred while accepting a new client: %s", err.Error()), ERROR_LEVEL)
				}
				continue
			}

			srv.Log(fmt.Sprintf("A new client - %s has connected to the server", TextColor.Green(clientConnection.RemoteAddr().String())), INFO_LEVEL)
			srv.wg.Add(1)
			go srv.handleClient(clientConnection)
			srv.cw.UpdateCount(1)
		}
	}
}

// Handles incoming HTTP requests sent from each individual client trying to connect to the web server instance.
func (srv *HttpServer) handleClient(ClientConnection net.Conn) {
	defer srv.wg.Done()
	defer srv.cw.UpdateCount(-1)
	defer ClientConnection.Close()

	handleRequest := func() (int, error) {
		httpRequest := srv.NewRequest(ClientConnection)
		err := httpRequest.Read()
		if err != nil {
			_, ok := err.(*ReadTimeoutError)
			if err != io.EOF && !ok {
				srv.Log(err.Error(), ERROR_LEVEL)
			}

			return 0, err
		}

		httpRequest.Locals["Started"] = time.Now()
		httpResponse := srv.NewResponse(ClientConnection, httpRequest)
		var timeout int
		var max int
		connValue, ok := httpRequest.Headers.Get("Connection")
		if ok && strings.EqualFold(connValue, "keep-alive") && strings.EqualFold(httpResponse.Version, "1.1") {
			currCount := srv.cw.GetCount()
			timeout, max = srv.getKeepAliveHeuristic(currCount)
			srv.Log(fmt.Sprintf("The timeout value returned by heuristic is %d seconds for active connection count %d", timeout, currCount), INFO_LEVEL)
			tcpConn, ok := ClientConnection.(*net.TCPConn)
			if ok {
				tcpConn.SetKeepAlive(true)
				tcpConn.SetKeepAlivePeriod(time.Duration(timeout) * time.Second)
				tcpConn.SetReadDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
			}

			httpResponse.Headers.Add("Connection", "keep-alive")
			httpResponse.Headers.Add("Keep-Alive", fmt.Sprintf("timeout=%d, max=%d", timeout, max))
		}

		if !IsMethodAllowed(httpResponse.Version, strings.ToUpper(strings.TrimSpace(httpRequest.Method))) {
			httpResponse.Status(Status405)
			ErrorHandler(httpRequest, httpResponse)
		} else {
			// First stage of execution will implement all server level middlewares configured.
			if len(srv.middlewares) > 0 {
				responseSent := srv.processMiddlewares(httpRequest, httpResponse, srv.middlewares)
				if responseSent {
					srv.logStatus(httpRequest, httpResponse)
					return timeout, nil
				}
			}

			// Next match the request route with the route tree to find the corresponding route handler.
			matchedRoute, err := srv.Router.Match(httpRequest)
			if err != nil {
				srv.Log(err.Error(), ERROR_LEVEL)
				httpResponse.Status(Status404)
				ErrorHandler(httpRequest, httpResponse)
			} else {
				// After match is fetched, process the route level middlewares.
				if len(matchedRoute.Middlewares) > 0 {
					responseSent := srv.processMiddlewares(httpRequest, httpResponse, matchedRoute.Middlewares)
					if responseSent {
						srv.logStatus(httpRequest, httpResponse)
						return timeout, nil
					}
				}

				// Finally execute the handler for the matched route.
				matchedRoute.RouteHandler(httpRequest, httpResponse)
			}
		}

		srv.logStatus(httpRequest, httpResponse)
		return timeout, nil
	}

	timer := time.NewTimer(0)
	defer timer.Stop()

	for {
		timeout, err := handleRequest()
		_, ok := err.(*ReadTimeoutError)
		if err != io.EOF && !ok {
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(time.Duration(timeout) * time.Second)
			srv.Log(fmt.Sprintf("Timeout for client connection [%s] has been changed to %d seconds.", ClientConnection.RemoteAddr().String(), timeout), WARN_LEVEL)
		}

		select {
		case <-srv.shutdown:
			srv.Log("Server shutdown initiated :: Closing client connection - "+ClientConnection.RemoteAddr().String(), WARN_LEVEL)
			return
		case <-timer.C:
			srv.Log(fmt.Sprintf("Client connection [%s] has timed out.", ClientConnection.RemoteAddr().String()), INFO_LEVEL)
			return
		default:
		}
	}
}

// Server's Keep-Alive heuristic which returns the timeout value and the maximum number of requests that can be processed by a single connection.
func (srv *HttpServer) getKeepAliveHeuristic(connCount int) (int, int) {
	usableCPU := numCPU - 1
	scalingFactor := 2.0
	timeout := 15 / (1 + math.Exp(scalingFactor*float64(connCount-usableCPU)))
	return int(math.Ceil(timeout)), 100
}

// Terminate all the active connections with the server before shutting down the server instance.
func (srv *HttpServer) terminate() {
	srv.Log("Server shutdown signal received...", INFO_LEVEL)
	terminateDone := make(chan struct{})
	go func() {
		srv.Log("Server Shutdown :: All existing connections are being terminated.", WARN_LEVEL)
		close(srv.shutdown)
		srv.close()
		srv.wg.Wait()
		close(terminateDone)
	}()

	srvShutTimeout := GetServerDefaults("shutdown_timeout").(int)

	select {
	case <-terminateDone:
		srv.Log("Server Shutdown :: All active connections have been terminated successfully.", INFO_LEVEL)
		return
	case <-time.After(time.Duration(srvShutTimeout) * time.Second):
		srv.Log("Server Shutdown Timeout :: Not all active connection(s) were terminated successfully.", ERROR_LEVEL)
		return
	}
}

// Logs the status of a request once its processing is completed.
// The details entered in the log dependes on the log format configured for the server.
//
// "common" [default log format] << :remote-addr [:date[clf]] ":method :url HTTP/:http-version" :status :res[content-length] >> This follows the common Apache log format
//
// "short" << :remote-addr :method :url HTTP/:http-version :status :res[content-length] - :response-time ms >>
//
// "tiny" << :method :url :status :res[content-length] - :response-time ms >>
//
// "dev" << :method :url :status :response-time ms - :res[content-length] >>
func (srv *HttpServer) logStatus(request *HttpRequest, response *HttpResponse) {
	reqProcessingTime := request.ProcessingTime()
	if strings.EqualFold(srv.logFormat, DEV_LOGGER) {
		responseContentLength, ok := response.Headers.Get("Content-Length")
		if !ok {
			responseContentLength = "-"
		}
		statusString := ""
		if response.StatusCode >= 200 && response.StatusCode < 300 {
			statusString = TextColor.Green(fmt.Sprintf("%d", response.StatusCode))
		} else if response.StatusCode >= 300 && response.StatusCode < 400 {
			statusString = TextColor.Cyan(fmt.Sprintf("%d", response.StatusCode))
		} else if response.StatusCode >= 400 && response.StatusCode < 500 {
			statusString = TextColor.Yellow(fmt.Sprintf("%d", response.StatusCode))
		} else if response.StatusCode >= 500 {
			statusString = TextColor.Red(fmt.Sprintf("%d", response.StatusCode))
		}
		srv.requestLogger.Printf("%s %s %s %d ms - %s", request.Method, request.ResourcePath, statusString, reqProcessingTime, responseContentLength)
	} else if strings.EqualFold(srv.logFormat, TINY_LOGGER) {
		responseContentLength, ok := response.Headers.Get("Content-Length")
		if !ok {
			responseContentLength = "-"
		}
		srv.requestLogger.Printf("%s %s %d %s - %d ms", request.Method, request.ResourcePath, response.StatusCode, responseContentLength, reqProcessingTime)
	} else if strings.EqualFold(srv.logFormat, SHORT_LOGGER) {
		responseContentLength, ok := response.Headers.Get("Content-Length")
		if !ok {
			responseContentLength = "-"
		}
		srv.requestLogger.Printf("%s %s %s HTTP/%s %d %s - %d ms", request.ClientAddress, request.Method, request.ResourcePath, request.Version, response.StatusCode, responseContentLength, reqProcessingTime)
	} else {
		responseContentLength, ok := response.Headers.Get("Content-Length")
		if !ok {
			responseContentLength = "-"
		}
		srv.requestLogger.Printf("%s %s \"%s %s HTTP/%s\" %d %s", request.ClientAddress, GetCLFTime(), request.Method, request.ResourcePath, request.Version, response.StatusCode, responseContentLength)
	}
}

// NewRequest creates and initializes a new HttpRequest for the given connection.
// Sets up request with default values, client address, and server reference.
// Called internally by the server for each incoming connection.
// Returns pointer to the initialized HttpRequest instance.
func (srv *HttpServer) NewRequest(Connection net.Conn) *HttpRequest {
	var httpRequest HttpRequest
	httpRequest.Initialize(Connection)
	httpRequest.ClientAddress = Connection.RemoteAddr().String()
	httpRequest.Server = srv
	return &httpRequest
}

// NewResponse creates and initializes a new HttpResponse for the given connection.
// Sets up response with appropriate HTTP version, default headers, and server reference.
// Called internally by the server for each request.
// Returns pointer to the initialized HttpResponse instance.
func (srv *HttpServer) NewResponse(Connection net.Conn, request *HttpRequest) *HttpResponse {
	var httpResponse HttpResponse
	httpResponse.Initialize(GetResponseVersion(request.Version), Connection)
	httpResponse.Server = srv
	return &httpResponse
}

// SetLogger configures the logging format for HTTP request processing logs.
// Supports COMMON_LOGGER, DEV_LOGGER, TINY_LOGGER, and SHORT_LOGGER formats.
// Defaults to COMMON_LOGGER if an unsupported format is provided.
// Use to customize how requests and responses are logged to console.
func (srv *HttpServer) SetLogger(logFormat string) {
	allowedLogFormats := []string{COMMON_LOGGER, DEV_LOGGER, TINY_LOGGER, SHORT_LOGGER}
	if slices.Contains(allowedLogFormats, logFormat) {
		srv.logFormat = logFormat
		srv.Log(fmt.Sprintf("The request logging format has been set to '%s'.", logFormat), WARN_LEVEL)
	} else {
		srv.logFormat = COMMON_LOGGER
		srv.Log(fmt.Sprintf("The log format value given was not valid so the default value - %s has been setup.", COMMON_LOGGER), ERROR_LEVEL)
	}
}

// Use registers a middleware function to execute for all incoming HTTP requests.
// Server-level middleware runs before route-specific middleware and handlers.
// Middleware functions execute in the order they were registered.
// Use for authentication, logging, body parsing, CORS, etc.
func (srv *HttpServer) Use(middleware Middleware) {
	srv.middlewares = append(srv.middlewares, middleware)
}

// Listen starts the HTTP server and begins accepting connections.
// Blocks until terminated by interrupt signal (Ctrl+C, SIGINT, SIGTERM).
// Creates TCP listener, handles graceful shutdown, and manages connections.
// Ensure HostAddress, PortNumber, and Router are configured before calling.
func (srv *HttpServer) Listen() {
	serverAddress := fmt.Sprintf("%s:%d", srv.HostAddress, srv.PortNumber)
	server, err := net.Listen("tcp", serverAddress)
	if err != nil {
		srv.Log(fmt.Sprintf("Error occurred while setting up listener socket: %s", err.Error()), ERROR_LEVEL)
		return
	}

	srv.listener = server
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	if srv.shutdown == nil {
		srv.shutdown = make(chan struct{})
	}

	srv.cw = new(ConnectionWatcher)
	srv.Log(fmt.Sprintf("Web server is listening at http://%s", serverAddress), WARN_LEVEL)
	srv.Log("To terminate the server, press Ctrl + C", WARN_LEVEL)
	srv.wg.Add(1)
	go srv.acceptConnections()
	<-sigChan
	srv.terminate()
	close(sigChan)
}

// Log outputs a message to the server's log stream with specified log level.
// Supports INFO_LEVEL, WARN_LEVEL (yellow), and ERROR_LEVEL (red) with colors.
// Includes timestamp, server name, log level, and message in output.
// Thread-safe method used for operational logging by server and applications.
func (srv *HttpServer) Log(message string, level string) {
	currentTime := GetRfc1123Time()
	serverName := GetServerDefaults("server_name").(string)
	if level == ERROR_LEVEL {
		srv.requestLogger.Printf(TextColor.Red("%s %s %s %s"), currentTime, serverName, strings.ToUpper(strings.TrimSpace(level)), strings.TrimSpace(message))
	} else if level == WARN_LEVEL {
		srv.requestLogger.Printf(TextColor.Yellow("%s %s %s %s"), currentTime, serverName, strings.ToUpper(strings.TrimSpace(level)), strings.TrimSpace(message))
	} else {
		srv.requestLogger.Printf("%s %s %s %s", currentTime, serverName, strings.ToUpper(strings.TrimSpace(level)), strings.TrimSpace(message))
	}
}
