package job

import (
	"encoding/json"
	"testing"
	"time"
)

func TestStatus_constants(t *testing.T) {
	cases := []Status{
		StatusPending, StatusRunning, StatusDone, StatusFailed, StatusCancelled,
	}
	for _, c := range cases {
		if string(c) == "" {
			t.Fatalf("empty Status constant")
		}
	}
}

func TestJob_JSONRoundTrip(t *testing.T) {
	created := time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC)
	updated := created.Add(5 * time.Minute)
	in := Job{
		ID:         "job-1",
		ToolName:   "nmap_scan",
		Target:     "10.0.0.1",
		Subject:    "engage.scan",
		Parameters: map[string]string{"ports": "443"},
		Status:     StatusRunning,
		Output:     "partial",
		CreatedAt:  created,
		UpdatedAt:  updated,
	}
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var out Job
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if out.ID != in.ID || out.Status != in.Status || out.Parameters["ports"] != "443" {
		t.Fatalf("got %+v", out)
	}
	if !out.CreatedAt.Equal(created) || !out.UpdatedAt.Equal(updated) {
		t.Fatalf("times got created=%v updated=%v", out.CreatedAt, out.UpdatedAt)
	}
}

func TestJob_zeroSafe(t *testing.T) {
	var j Job
	if j.ID != "" || j.ToolName != "" || j.Status != "" || j.Parameters != nil {
		t.Fatal("zero Job should be empty")
	}
}
