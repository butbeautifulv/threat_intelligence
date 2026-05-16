package intelligence

import (
	"context"
	"testing"
)

func TestProbeTarget_urlWeb(t *testing.T) {
	tt, tech, cms, conf, _, _ := probeTarget(context.Background(), "https://example.com/wp-admin")
	if tt != "web" {
		t.Fatalf("target type: %s", tt)
	}
	if cms != "wordpress" {
		t.Fatalf("cms: %q", cms)
	}
	if conf < 0.5 {
		t.Fatalf("confidence: %f", conf)
	}
	if len(tech) == 0 {
		t.Fatal("expected technologies")
	}
}

func TestProbeTarget_ip(t *testing.T) {
	tt, _, _, conf, _, _ := probeTarget(context.Background(), "192.168.1.1")
	if tt != "ip" {
		t.Fatalf("target type: %s", tt)
	}
	if conf < 0.7 {
		t.Fatalf("confidence: %f", conf)
	}
}

func TestSelectTools_quickObjective(t *testing.T) {
	s := &Service{
		Registry: testRegistry(webCatalogSpecs()),
		Engine:   DefaultDecisionEngine(),
	}
	got := s.SelectTools(context.Background(), "web", "quick")
	if len(got) > 3 {
		t.Fatalf("quick should cap at 3, got %v", got)
	}
}

func TestTechnologiesDetected_labels(t *testing.T) {
	labels := technologiesDetected([]string{"nginx/1.18", "PHP/8.1"}, "wordpress")
	found := map[string]bool{}
	for _, l := range labels {
		found[l] = true
	}
	if !found["wordpress"] || !found["nginx"] || !found["php"] {
		t.Fatalf("labels: %v", labels)
	}
}

func TestTechnologyDetection_shape(t *testing.T) {
	s := &Service{Engine: DefaultDecisionEngine()}
	out := s.TechnologyDetection(context.Background(), "https://example.com")
	if out["technologies"] == nil {
		t.Fatal("missing technologies")
	}
	if _, ok := out["confidence"]; !ok {
		t.Fatal("missing confidence")
	}
}
