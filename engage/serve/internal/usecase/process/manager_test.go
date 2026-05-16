package process

import "testing"

func TestDashboard_enriched(t *testing.T) {
	m := NewManager()
	m.Register(1001, "nmap_scan", "127.0.0.1", "nmap -sn 127.0.0.1")
	m.UpdateProgress(1001, 0.5, "scanning", 100)
	d := m.Dashboard()
	if d["system_load"] == nil {
		t.Fatal("missing system_load")
	}
	if d["timestamp"] == nil {
		t.Fatal("missing timestamp")
	}
	procs, ok := d["processes"].([]map[string]any)
	if !ok || len(procs) == 0 {
		t.Fatalf("processes %#v", d["processes"])
	}
	if procs[0]["progress_percent"] == nil {
		t.Fatal("missing progress_percent")
	}
}
