package internal

import (
	"path/filepath"
	"strings"
)

// Structure to contain information about a single route declared in the Router.
type Route struct {
	// Handler function to be executed for the route paths.
	RouteHandler RouteHandler
	// HTTP method for which the route is defined
	Method string
	// List of all route level middlewares configured.
	Middlewares []Middleware
}

// Structure to hold all the routes and the associated routing logic.
type Router struct {
	// Prefix tree containing all the routes declared on the router.
	routeTree *PrefixTree
	// Collection of static paths configured for the router.
	staticRoutes map[string]string
	// To access the underlying filesystem and its files/folders.
	fs *FileSystem
}

// Static registers a route to serve static files from specified filesystem directory.
// Use to serve CSS, JS, images, and other assets with automatic MIME detection and caching.
// Maps URL prefix to filesystem path with security validation and directory traversal protection.
// Parameters: RoutePath (URL prefix), TargetPath (absolute filesystem directory).
// Returns: error if target path invalid or inaccessible, nil on successful registration.
func (rtr *Router) Static(RoutePath string, TargetPath string) error {
	RoutePath = CleanRoute(RoutePath)
	isAbsolute := rtr.fs.IsAbsolute(TargetPath)
	if !isAbsolute {
		reError := new(RoutingError)
		reError.RoutePath = TargetPath
		reError.Message = "Target path must be absolute"
		return reError
	}
	isDirectory := rtr.fs.IsDirectory(TargetPath)
	if !isDirectory {
		reError := new(RoutingError)
		reError.RoutePath = TargetPath
		reError.Message = "Target path given should point to a directory in the local file system"
		return reError
	}

	rtr.staticRoutes[RoutePath] = TargetPath
	return nil
}

// Get registers HTTP GET endpoint with handler and optional middleware for data retrieval.
// Use for web pages, API endpoints, content serving with static/dynamic path support.
// Supports path parameters (:id), wildcards (*), and middleware chain execution.
// Parameters: RoutePath (URL pattern), handlerFunc (route handler), middlewareList (optional middleware).
// Returns: error if route registration fails, nil on successful registration.
func (rtr *Router) Get(RoutePath string, handlerFunc RouteHandler, middlewareList ...Middleware) error {
	RoutePath = CleanRoute(RoutePath)
	return rtr.addRoute("GET", RoutePath, handlerFunc, middlewareList)
}

// Head registers HTTP HEAD endpoint for retrieving headers without body content.
// Use for resource validation, metadata checking, and bandwidth-efficient probing.
// Returns same headers as GET but no response body for efficient resource validation.
// Parameters: RoutePath (URL pattern), handlerFunc (route handler), middlewareList (optional middleware).
// Returns: error if route registration fails, nil on successful registration.
func (rtr *Router) Head(RoutePath string, handlerFunc RouteHandler, middlewareList ...Middleware) error {
	RoutePath = CleanRoute(RoutePath)
	return rtr.addRoute("HEAD", RoutePath, handlerFunc, middlewareList)
}

// Post registers HTTP POST endpoint for creating resources and processing form/JSON data.
// Use for user creation, form submissions, file uploads, and API operations that modify state.
// Supports body parsing middleware (JsonParser, UrlEncoded) for request data processing.
// Parameters: RoutePath (URL pattern), handlerFunc (route handler), middlewareList (optional middleware).
// Returns: error if route registration fails, nil on successful registration.
func (rtr *Router) Post(RoutePath string, handlerFunc RouteHandler, middlewareList ...Middleware) error {
	RoutePath = CleanRoute(RoutePath)
	return rtr.addRoute("POST", RoutePath, handlerFunc, middlewareList)
}

// Put registers HTTP PUT endpoint for updating/replacing resources with idempotent operations.
// Use for complete resource updates, configuration changes, and file uploads to specific URLs.
// Idempotent behavior (safe to repeat), client specifies resource location unlike POST.
// Parameters: RoutePath (URL pattern with ID), handlerFunc (route handler), middlewareList (optional middleware).
// Returns: error if route registration fails, nil on successful registration.
func (rtr *Router) Put(RoutePath string, handlerFunc RouteHandler, middlewareList ...Middleware) error {
	RoutePath = CleanRoute(RoutePath)
	return rtr.addRoute("PUT", RoutePath, handlerFunc, middlewareList)
}

// Delete registers HTTP DELETE endpoint for removing resources permanently with idempotent behavior.
// Use for resource deletion, cleanup operations, session logout, and administrative management.
// Returns 204 No Content on success, handles non-existent resources gracefully.
// Parameters: RoutePath (URL pattern with ID), handlerFunc (route handler), middlewareList (optional middleware).
// Returns: error if route registration fails, nil on successful registration.
func (rtr *Router) Delete(RoutePath string, handlerFunc RouteHandler, middlewareList ...Middleware) error {
	RoutePath = CleanRoute(RoutePath)
	return rtr.addRoute("DELETE", RoutePath, handlerFunc, middlewareList)
}

// Trace registers HTTP TRACE endpoint for diagnostic loopback testing and request tracing.
// Use for debugging HTTP request path, proxy behavior, and network diagnostics.
// Echoes received request headers back to client for troubleshooting purposes.
// Parameters: RoutePath (URL pattern), handlerFunc (route handler), middlewareList (optional middleware).
// Returns: error if route registration fails, nil on successful registration.
func (rtr *Router) Trace(RoutePath string, handlerFunc RouteHandler, middlewareList ...Middleware) error {
	RoutePath = CleanRoute(RoutePath)
	return rtr.addRoute("TRACE", RoutePath, handlerFunc, middlewareList)
}

// Options registers HTTP OPTIONS endpoint for CORS preflight requests and method discovery.
// Use for cross-origin API access, capability discovery, and browser preflight handling.
// Sets CORS headers (Allow-Origin, Allow-Methods, Allow-Headers) for secure cross-origin requests.
// Parameters: RoutePath (URL pattern), handlerFunc (route handler), middlewareList (optional middleware).
// Returns: error if route registration fails, nil on successful registration.
func (rtr *Router) Options(RoutePath string, handlerFunc RouteHandler, middlewareList ...Middleware) error {
	RoutePath = CleanRoute(RoutePath)
	return rtr.addRoute("OPTIONS", RoutePath, handlerFunc, middlewareList)
}

// Connect registers HTTP CONNECT endpoint for establishing tunnels through proxy servers.
// Use for proxy connections, SSL tunneling, and WebSocket upgrade handling.
// Typically used by proxy servers and clients requiring secure tunnel establishment.
// Parameters: RoutePath (URL pattern), handlerFunc (route handler), middlewareList (optional middleware).
// Returns: error if route registration fails, nil on successful registration.
func (rtr *Router) Connect(RoutePath string, handlerFunc RouteHandler, middlewareList ...Middleware) error {
	RoutePath = CleanRoute(RoutePath)
	return rtr.addRoute("CONNECT", RoutePath, handlerFunc, middlewareList)
}

// addRoute adds dynamic route with handler and middleware to router's prefix tree.
// Use internally by HTTP method functions (Get, Post, etc.) to register routes.
// Validates method, cleans path, creates route object, and inserts into prefix tree.
// Parameters: Method (HTTP method), RoutePath (URL pattern), handlerFunc (handler), middlewareList (middleware).
// Returns: error if route registration fails, nil on successful addition.
func (rtr *Router) addRoute(Method string, RoutePath string, handlerFunc RouteHandler, middlewareList []Middleware) error {
	RoutePath = CleanRoute(RoutePath)
	Method = strings.TrimSpace(Method)
	Method = strings.ToUpper(Method)

	routeObj := Route{
		RouteHandler: handlerFunc,
		Method:       Method,
		Middlewares:  make([]Middleware, 0),
	}

	routeObj.Middlewares = append(routeObj.Middlewares, middlewareList...)
	rtr.routeTree.Insert(RoutePath, &routeObj)
	return nil
}

// Match finds route handler for HTTP request by searching static files and dynamic patterns.
// Use internally by server to locate appropriate handler for incoming requests.
// Checks static routes first (GET/HEAD), then dynamic patterns with parameter extraction.
// Parameter: request (HttpRequest with path and method to match).
// Returns: *Route with handler and middleware, error if no match found.
func (rtr *Router) Match(request *HttpRequest) (*Route, error) {
	routePath := CleanRoute(request.ResourcePath)
	if strings.EqualFold(request.Method, "GET") || strings.EqualFold(request.Method, "HEAD") {
		for routeKey, TargetPath := range rtr.staticRoutes {
			if strings.HasPrefix(routePath, routeKey) {
				RouteAfterPrefix := strings.TrimPrefix(routePath, routeKey)
				RouteAfterPrefix = CleanRoute(RouteAfterPrefix)
				FinalPath := filepath.Join(TargetPath, RouteAfterPrefix)
				if rtr.fs.Exists(FinalPath) {
					request.Locals["StaticFilePath"] = FinalPath
					finalRoute := new(Route)
					finalRoute.Method = request.Method
					finalRoute.RouteHandler = StaticFileHandler
					finalRoute.Middlewares = make([]Middleware, 0)
					return finalRoute, nil
				}
			}
		}
	}

	routeInfo := rtr.routeTree.Match(routePath)
	if routeInfo.MatchedRoutes == nil {
		reError := new(RoutingError)
		reError.RoutePath = routePath
		reError.Message = "matchRoute: A match was not found in the router's prefix tree"
		return nil, reError
	}

	if routeInfo.Segments.Length() > 0 {
		for key, values := range routeInfo.Segments {
			request.Segments.Add(key, values)
		}
	}

	var finalRoute *Route = nil
	for _, route := range routeInfo.MatchedRoutes {
		if strings.EqualFold(route.Method, request.Method) {
			finalRoute = route
			break
		}
	}

	if finalRoute == nil {
		reError := new(RoutingError)
		reError.RoutePath = routePath
		reError.Message = "matchRoute: A match was not for the HTTP method and route combination"
		return nil, reError
	}

	return finalRoute, nil
}

// Creates a new instance of Router and returns a reference to the instance.
func NewRouter() *Router {
	router := new(Router)
	router.routeTree = EmptyPrefixTree()
	router.staticRoutes = make(map[string]string)
	router.fs = new(FileSystem)
	return router
}
