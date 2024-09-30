package http

import (
	"net"
	"strconv"
	"strings"
	"fmt"
)

type HttpServer struct {
	HostAddress string
	PortNumber int
	Socket net.Listener
	StaticRouter FileRoutes
}

func (srv *HttpServer) Static(Route string, TargetPath string) {
	if srv.StaticRouter == nil {
		srv.StaticRouter = make(FileRoutes)
	}
	err := srv.StaticRouter.Add(Route, TargetPath)
	if err != nil {
		LogError(fmt.Sprintf("AddStaticRoute() :: %s", err.Error()))
		return
	}
}

func (srv * HttpServer) Listen(PortNumber int, HostAddress string) {
	if PortNumber == 0 {
		srv.PortNumber = GetDefaultPort()
	} else {
		srv.PortNumber = PortNumber
	}

	if HostAddress == "" {
		srv.HostAddress = GetDefaultHostname()
	} else {
		srv.HostAddress = strings.TrimSpace(HostAddress)
	}

	serverAddress := srv.HostAddress + ":" + strconv.Itoa(srv.PortNumber)
	server, err := net.Listen("tcp", serverAddress)
	if err != nil {
		LogError(fmt.Sprintf("Error occurred while setting up listener socket: %s", err.Error()))
		return
	}
	srv.Socket = server
	defer srv.Socket.Close()
	LogError(fmt.Sprintf("Web server is listening at http://%s", serverAddress))

	for {
		clientConnection, err := srv.Socket.Accept()
		if err != nil {
			LogError(fmt.Sprintf("Error occurred while accepting a new client: %s", err.Error()))
			continue
		}
		LogError(fmt.Sprintf("A new client - %s has connected to the server", clientConnection.RemoteAddr().String()))
		go srv.handleClient(clientConnection)
	}
}

func (srv *HttpServer) handleClient(ClientConnection net.Conn) {
	defer ClientConnection.Close()
	httpRequest := NewRequest(ClientConnection)
	httpRequest.read()
	httpResponse := NewResponse(ClientConnection, httpRequest)
	
	switch httpRequest.Method {
	case "GET":
		if !IsMethodAllowed(httpResponse.Version, "GET") {
			LogError("'GET' method is not allowed in HTTP version " + httpResponse.Version)
			httpResponse.Status(StatusMethodNotAllowed)
			httpResponse.AddHeader("Allow", GetAllowedMethods(httpResponse.Version))
			httpResponse.SendError(StatusMethodNotAllowed.GetErrorContent())
			return
		}
		TargetFilePath, staticRouteExists := srv.StaticRouter.Match(httpRequest.ResourcePath)
		if !staticRouteExists {
			LogError("A Static route matching the given resource does  not exist")
			httpResponse.Status(StatusNotFound)
			httpResponse.SendError(StatusNotFound.GetErrorContent())
			return
		}

		if !httpRequest.isConditionalGet(TargetFilePath) {
			httpResponse.Status(StatusOK)
			httpResponse.SendFile(TargetFilePath, false)
		} else {
			httpResponse.Status(StatusNotModified)
			httpResponse.SendFile(TargetFilePath, true)
		}
	case "HEAD":
		if !IsMethodAllowed(httpResponse.Version, "HEAD") {
			LogError("'HEAD' method is not allowed in HTTP version " + httpResponse.Version)
			httpResponse.Status(StatusMethodNotAllowed)
			httpResponse.AddHeader("Allow", GetAllowedMethods(httpResponse.Version))
			httpResponse.SendError(StatusMethodNotAllowed.GetErrorContent())
			return
		}
		TargetFilePath, staticRouteExists := srv.StaticRouter.Match(httpRequest.ResourcePath)
		if !staticRouteExists {
			LogError("A Static route matching the given resource does  not exist")
			httpResponse.Status(StatusNotFound)
			httpResponse.SendError(StatusNotFound.GetErrorContent())
			return
		}

		httpResponse.Status(StatusOK)
		httpResponse.SendFile(TargetFilePath, true)
	default:
		LogError("The HTTP method is not supported by the server. Allowed Methods are - " + GetAllowedMethods(httpResponse.Version))
		httpResponse.Status(StatusMethodNotAllowed)
		httpResponse.AddHeader("Allow", GetAllowedMethods(httpResponse.Version))
		httpResponse.SendError(StatusMethodNotAllowed.GetErrorContent())
	}
}