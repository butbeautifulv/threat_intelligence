package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
)

func TestFramedRW_roundTrip(t *testing.T) {
	var buf bytes.Buffer
	rw := NewFramedRW(&buf, &buf)
	ctx := context.Background()

	in := Message{JSONRPC: "2.0", ID: "req-7", Method: "ping"}
	if err := rw.WriteJSON(ctx, in); err != nil {
		t.Fatal(err)
	}

	raw, err := rw.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var out Message
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	if out.JSONRPC != in.JSONRPC || out.ID != in.ID || out.Method != in.Method {
		t.Fatalf("header fields: %+v", out)
	}
}

func TestFramedRW_writeRead_multiple(t *testing.T) {
	var buf bytes.Buffer
	rw := NewFramedRW(&buf, &buf)
	ctx := context.Background()

	for i, id := range []any{"1", "2", "three"} {
		msg := Message{JSONRPC: "2.0", ID: id, Method: "echo"}
		if err := rw.WriteJSON(ctx, msg); err != nil {
			t.Fatalf("[%d] write: %v", i, err)
		}
		raw, err := rw.Read(ctx)
		if err != nil {
			t.Fatalf("[%d] read: %v", i, err)
		}
		var got Message
		if err := json.Unmarshal(raw, &got); err != nil {
			t.Fatalf("[%d] unmarshal: %v", i, err)
		}
		if got.ID != id {
			t.Fatalf("[%d] id: got %v want %v", i, got.ID, id)
		}
	}
}
