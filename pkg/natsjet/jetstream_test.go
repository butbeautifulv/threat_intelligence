package natsjet

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestConnect_badURL(t *testing.T) {
	t.Parallel()
	_, err := Connect("not-a-valid-nats-url://")
	if err == nil {
		t.Fatal("expected connect error")
	}
}

func TestConnect_ok(t *testing.T) {
	c := connForTest(t)
	url := c.NC.ConnectedUrl()
	c.Close()
	got, err := Connect(url)
	if err != nil {
		t.Fatal(err)
	}
	defer got.Close()
	if got.NC == nil || got.JS == nil {
		t.Fatal("expected NC and JS")
	}
}

func TestPublishJSON_msgIDDedup(t *testing.T) {
	c := connForTest(t)
	if err := EnsureStream(c.JS, "W6_TEST", []string{"w6.test.>"}); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	payload := map[string]string{"id": "1"}
	msgID := "w6:dedup:1"
	if err := c.PublishJSON(ctx, "w6.test.events", payload, msgID); err != nil {
		t.Fatal(err)
	}
	if err := c.PublishJSON(ctx, "w6.test.events", payload, msgID); err != nil {
		t.Fatal(err)
	}
	info, err := c.JS.StreamInfo("W6_TEST")
	if err != nil {
		t.Fatal(err)
	}
	if info.State.Msgs != 1 {
		t.Fatalf("stream msgs = %d, want 1 (Nats-Msg-Id dedup)", info.State.Msgs)
	}
}

func TestPublishJSON_errors(t *testing.T) {
	c := connForTest(t)
	if err := EnsureStream(c.JS, "W6_ERR", []string{"w6.err.>"}); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	err := c.PublishJSON(ctx, "w6.err.events", make(chan int), "id-1")
	if err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("marshal err = %v", err)
	}

	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	err = c.PublishJSON(cancelled, "w6.err.events", map[string]string{"k": "v"}, "id-2")
	if err != context.Canceled {
		t.Fatalf("ctx err = %v", err)
	}
}

func TestEnsureStream_idempotent(t *testing.T) {
	js := jetStreamForTest(t)
	subjects := []string{"events.w6.>"}
	if err := EnsureStream(js, "W6_IDEM", subjects); err != nil {
		t.Fatal(err)
	}
	if err := EnsureStream(js, "W6_IDEM", subjects); err != nil {
		t.Fatal(err)
	}
	info, err := js.StreamInfo("W6_IDEM")
	if err != nil {
		t.Fatal(err)
	}
	if len(info.Config.Subjects) != 1 || info.Config.Subjects[0] != subjects[0] {
		t.Fatalf("subjects %#v", info.Config.Subjects)
	}
}

func TestConn_Close(t *testing.T) {
	c := connForTest(t)
	c.Close()
	deadline := time.Now().Add(2 * time.Second)
	for c.NC.IsConnected() {
		if time.Now().After(deadline) {
			t.Fatal("connection still open after Close")
		}
		time.Sleep(10 * time.Millisecond)
	}
}
