package bugbounty

import "testing"

func TestCreateReconnaissance_phases(t *testing.T) {
	m := NewManager()
	wf := m.CreateReconnaissance(Target{Domain: "example.com"})
	if len(wf.Phases) < 4 {
		t.Fatalf("phases %d", len(wf.Phases))
	}
	if wf.ToolsCount < 8 {
		t.Fatalf("tools_count %d", wf.ToolsCount)
	}
	if wf.EstimatedTime <= 0 {
		t.Fatal("estimated_time")
	}
	if wf.Phases[0].Name != "subdomain_discovery" {
		t.Fatalf("first phase %q", wf.Phases[0].Name)
	}
}

func TestCreateVulnHunt_priorityOrder(t *testing.T) {
	m := NewManager()
	wf := m.CreateVulnHunt(Target{
		Domain:        "example.com",
		PriorityVulns: []string{"xss", "rce", "sqli"},
	})
	if len(wf.VulnerabilityTests) < 3 {
		t.Fatalf("tests %d", len(wf.VulnerabilityTests))
	}
	if wf.VulnerabilityTests[0].VulnerabilityType != "rce" {
		t.Fatalf("first %q", wf.VulnerabilityTests[0].VulnerabilityType)
	}
}

func TestCreateComprehensive_summary(t *testing.T) {
	m := NewManager()
	a := m.CreateComprehensive(Target{Domain: "example.com", IncludeOSINT: true, IncludeBusiness: true})
	if a.Summary["total_estimated_time"].(int) <= 0 {
		t.Fatal("summary time")
	}
}
