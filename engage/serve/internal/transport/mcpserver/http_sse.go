package mcpserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func wantsSSE(r *http.Request) bool {
	accept := strings.ToLower(r.Header.Get("Accept"))
	return strings.Contains(accept, "text/event-stream")
}

func writeSSEMessages(w http.ResponseWriter, messages []rpcMessage) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming not supported")
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	for _, msg := range messages {
		b, err := json.Marshal(msg)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "event: message\ndata: %s\n\n", b); err != nil {
			return err
		}
		flusher.Flush()
	}
	return nil
}
