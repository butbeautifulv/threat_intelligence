package tooldispatch

import "fmt"

// DispatchError is returned from Dispatcher for HTTP/MCP mapping.
type DispatchError struct {
	NotFound bool
	Message  string
}

func (e *DispatchError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func dispatchNotFound(format string, args ...any) error {
	return &DispatchError{NotFound: true, Message: fmt.Sprintf(format, args...)}
}

func dispatchToolError(format string, args ...any) error {
	return &DispatchError{Message: fmt.Sprintf(format, args...)}
}
