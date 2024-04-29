package http

const (
	SERVER_TYPE = "tcp"
	SERVER_NAME = "sparrow"
	ERROR_MSG_CONTENT_TYPE = "text/html"
	ALLOWED_METHODS = "GET, HEAD, POST"
	HEADER_LINE_SEPERATOR = "\r\n"
	REQUEST_LINE_SEPERATOR = " "
	HEADER_KEY_VALUE_SEPERATOR = ":"
	MAX_VERSION = "HTTP/1.0"
	GET_METHOD = "GET"
	POST_METHOD = "POST"
	HEAD_METHOD = "HEAD"
)

var COMPATIBLE_VERSIONS = []string {"HTTP/1.0"}