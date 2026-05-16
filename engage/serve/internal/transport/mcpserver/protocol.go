package mcpserver

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
)

type rpcMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type framedRW struct {
	r *bufio.Reader
	w io.Writer
	m sync.Mutex
}

func newFramedRW(r io.Reader, w io.Writer) *framedRW {
	return &framedRW{r: bufio.NewReader(r), w: w}
}

func (rw *framedRW) read(ctx context.Context) ([]byte, error) {
	var contentLen int
	for {
		line, err := rw.r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		k := strings.TrimSpace(strings.ToLower(parts[0]))
		v := strings.TrimSpace(parts[1])
		if k == "content-length" {
			n, _ := strconv.Atoi(v)
			contentLen = n
		}
	}
	if contentLen <= 0 {
		return nil, fmt.Errorf("missing/invalid Content-Length")
	}
	buf := make([]byte, contentLen)
	if _, err := io.ReadFull(rw.r, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func (rw *framedRW) writeJSON(ctx context.Context, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	rw.m.Lock()
	defer rw.m.Unlock()
	var out bytes.Buffer
	out.WriteString(fmt.Sprintf("Content-Length: %d\r\n\r\n", len(b)))
	out.Write(b)
	_, err = rw.w.Write(out.Bytes())
	return err
}
