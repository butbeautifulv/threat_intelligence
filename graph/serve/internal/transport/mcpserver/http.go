package mcpserver

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/butbeautifulv/veil/graph/serve/internal/config"
	"github.com/butbeautifulv/veil/graph/serve/internal/version"
)

// HTTPHandler serves Streamable HTTP MCP (POST JSON or SSE).
func HTTPHandler(s *Server, cfg config.MCPHTTPConfig) http.Handler {
	mux := http.NewServeMux()
	path := cfg.Path
	if path == "" {
		path = "/mcp"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":      true,
			"service": version.ServerName,
			"transport": "streamable-http",
		})
	})

	mux.HandleFunc("GET "+path, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("POST "+path, func(w http.ResponseWriter, r *http.Request) {
		s.serveMCPPost(w, r, cfg)
	})

	return mux
}

func (s *Server) serveMCPPost(w http.ResponseWriter, r *http.Request, cfg config.MCPHTTPConfig) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 4<<20))
	if err != nil {
		writeHTTPRPCError(w, http.StatusBadRequest, codeParseError, "read body failed")
		return
	}
	if len(body) == 0 {
		writeHTTPRPCError(w, http.StatusBadRequest, codeInvalidRequest, "empty body")
		return
	}

	msgs, err := parseInboundMessages(body)
	if err != nil {
		writeHTTPRPCError(w, http.StatusBadRequest, codeParseError, err.Error())
		return
	}

	var requests, notifications []rpcMessage
	for _, m := range msgs {
		if m.Method != "" && m.ID != nil {
			requests = append(requests, m)
		} else if m.Method != "" && m.ID == nil {
			notifications = append(notifications, m)
		}
	}

	// Client responses/notifications only → 202 Accepted.
	if len(requests) == 0 {
		for _, n := range notifications {
			_, _, _ = s.ProcessMessage(r.Context(), n, true)
		}
		w.WriteHeader(http.StatusAccepted)
		return
	}

	if len(msgs) > 1 {
		writeHTTPRPCError(w, http.StatusBadRequest, codeInvalidRequest, "batch not supported")
		return
	}

	var out []rpcMessage
	for _, req := range requests {
		for _, n := range notifications {
			_, _, _ = s.ProcessMessage(r.Context(), n, true)
		}
		resp, isNotification, perr := s.ProcessMessage(r.Context(), req, true)
		if perr != nil {
			writeHTTPRPCError(w, http.StatusInternalServerError, codeInternal, perr.Error())
			return
		}
		if isNotification || resp == nil {
			continue
		}
		out = append(out, *resp)
	}

	if len(out) == 0 {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	if wantsSSE(r, cfg.PreferSSE) {
		if err := writeSSEMessages(w, out); err != nil {
			s.logger.Error("mcp sse write failed", "err", err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if len(out) == 1 {
		_ = json.NewEncoder(w).Encode(out[0])
		return
	}
	_ = json.NewEncoder(w).Encode(out)
}

func parseInboundMessages(body []byte) ([]rpcMessage, error) {
	body = trimSpaceBytes(body)
	if len(body) > 0 && body[0] == '[' {
		var batch []rpcMessage
		if err := json.Unmarshal(body, &batch); err != nil {
			return nil, err
		}
		return batch, nil
	}
	var msg rpcMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		return nil, err
	}
	return []rpcMessage{msg}, nil
}

func trimSpaceBytes(b []byte) []byte {
	for len(b) > 0 && (b[0] == ' ' || b[0] == '\n' || b[0] == '\r' || b[0] == '\t') {
		b = b[1:]
	}
	for len(b) > 0 && (b[len(b)-1] == ' ' || b[len(b)-1] == '\n' || b[len(b)-1] == '\r' || b[len(b)-1] == '\t') {
		b = b[:len(b)-1]
	}
	return b
}

func writeHTTPRPCError(w http.ResponseWriter, status, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(rpcMessage{
		JSONRPC: "2.0",
		Error:   &rpcError{Code: code, Message: msg},
	})
}
