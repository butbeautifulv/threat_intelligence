package mcpserver

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/butbeautifulv/veil/engage/serve/internal/domain/tool"
	"github.com/butbeautifulv/veil/engage/serve/internal/runner"
	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
	toolsuc "github.com/butbeautifulv/veil/engage/serve/internal/usecase/tools"
	"github.com/butbeautifulv/veil/engage/serve/internal/version"
)

func testRunner() *toolsuc.Runner {
	specs := []tool.Spec{
		{Name: "echo_test", Category: "network", Binary: "echo", ArgsTemplate: []string{"{target}"}, TimeoutSec: 5, Description: "echo", Enabled: true},
	}
	reg := tools.NewRegistry(specs)
	return &toolsuc.Runner{Registry: reg, Exec: &runner.Executor{}}
}

func TestServer_initialize_tools_ping(t *testing.T) {
	srv := NewServer(testRunner(), nil, slog.Default())

	stdinR, stdinW := io.Pipe()
	stdoutR, stdoutW := io.Pipe()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Run(ctx, stdinR, stdoutW)
	}()

	writeReq := func(id int, method string, params any) {
		t.Helper()
		var rawParams json.RawMessage
		if params != nil {
			b, _ := json.Marshal(params)
			rawParams = b
		}
		rw := newFramedRW(strings.NewReader(""), stdinW)
		if err := rw.writeJSON(ctx, rpcMessage{
			JSONRPC: "2.0",
			ID:      id,
			Method:  method,
			Params:  rawParams,
		}); err != nil {
			t.Fatal(err)
		}
	}

	readResp := func() rpcMessage {
		t.Helper()
		rw := newFramedRW(stdoutR, io.Discard)
		payload, err := rw.read(ctx)
		if err != nil {
			t.Fatal(err)
		}
		var msg rpcMessage
		if err := json.Unmarshal(payload, &msg); err != nil {
			t.Fatal(err)
		}
		return msg
	}

	writeReq(1, "initialize", map[string]any{
		"protocolVersion": protocol20241105,
		"capabilities":    map[string]any{},
	})
	init := readResp()
	if init.Error != nil {
		t.Fatalf("initialize error: %+v", init.Error)
	}
	info, _ := init.Result.(map[string]any)["serverInfo"].(map[string]any)
	if info["name"] != version.ServerName {
		t.Fatalf("server name: %v", info["name"])
	}

	writeReq(2, "tools/list", nil)
	list := readResp()
	if list.Error != nil {
		t.Fatalf("tools/list error: %+v", list.Error)
	}
	toolsRes, _ := list.Result.(map[string]any)["tools"].([]any)
	if len(toolsRes) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(toolsRes))
	}

	writeReq(3, "ping", nil)
	ping := readResp()
	if ping.Error != nil {
		t.Fatalf("ping error: %+v", ping.Error)
	}

	cancel()
	_ = stdinW.Close()
	select {
	case err := <-errCh:
		if err != nil && err != context.Canceled && !strings.Contains(err.Error(), "closed") {
			t.Fatalf("run: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for server exit")
	}
}

func TestProcessMessage_tools_call(t *testing.T) {
	srv := NewServer(testRunner(), nil, slog.Default())
	params, _ := json.Marshal(map[string]any{
		"name":      "echo_test",
		"arguments": map[string]any{"target": "hello"},
	})
	resp, notif, err := srv.ProcessMessage(context.Background(), rpcMessage{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params:  params,
	}, false)
	if err != nil || notif || resp == nil || resp.Error != nil {
		t.Fatalf("call: notif=%v err=%v resp=%+v", notif, err, resp)
	}
}
