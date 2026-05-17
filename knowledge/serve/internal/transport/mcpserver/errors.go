package mcpserver

import "fmt"

const (
	codeParseError     = -32700
	codeInvalidRequest = -32600
	codeMethodNotFound = -32601
	codeInvalidParams  = -32602
	codeInternal       = -32603
	codeToolError      = -32000
	codeAuthError      = -32001
)

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *rpcError) Error() string { return e.Message }

func rpcErr(code int, msg string) error {
	return &rpcError{Code: code, Message: msg}
}

func rpcErrf(code int, format string, args ...any) error {
	return rpcErr(code, fmt.Sprintf(format, args...))
}
