package visual

import "testing"

func TestStore_scanProgress(t *testing.T) {
	s := NewStore()
	id := s.Create("https://example.com", "smart_scan", []string{"nmap_scan", "httpx_probe"})
	s.StartTool(id, "nmap_scan")
	s.CompleteTool(id, "nmap_scan", "success", 1.2)
	sp, ok := s.Get(id)
	if !ok {
		t.Fatal("scan not found")
	}
	if sp.ToolsCompleted != 1 {
		t.Fatalf("completed %d", sp.ToolsCompleted)
	}
	if sp.ProgressPercent < 40 {
		t.Fatalf("progress %v", sp.ProgressPercent)
	}
	s.Finish(id, "completed")
	sp, _ = s.Get(id)
	if sp.Status != "completed" {
		t.Fatalf("status %s", sp.Status)
	}
}
