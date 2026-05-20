package decision

import (
	"reflect"
	"testing"
)

func TestCapTools(t *testing.T) {
	names := []string{"a", "b", "c", "d", "e", "f"}
	tests := []struct {
		name      string
		objective string
		want      []string
	}{
		{"quick caps at 3", "quick", names[:3]},
		{"fast alias", "fast", names[:3]},
		{"focused caps at 5", "focused", names[:5]},
		{"stealth caps at 4", "stealth", names[:4]},
		{"unknown objective unchanged", "balanced", names},
		{"trim and case", "  STEALTH ", names[:4]},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := names
			if tt.name == "under cap unchanged" {
				in = []string{"a", "b"}
			}
			got := CapTools(in, tt.objective)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("CapTools(%q) = %v, want %v", tt.objective, got, tt.want)
			}
		})
	}
}

func TestCapToolsWithEngine(t *testing.T) {
	eng := DefaultDecisionEngine()
	display := []string{"Nmap", "Gobuster", "Nuclei", "Nikto", "SQLMap", "FFuf", "Feroxbuster"}
	shortID := func(s string) string {
		switch s {
		case "Nmap":
			return "nmap"
		case "Gobuster":
			return "gobuster"
		case "Nuclei":
			return "nuclei"
		case "Nikto":
			return "nikto"
		case "SQLMap":
			return "sqlmap"
		case "FFuf":
			return "ffuf"
		case "Feroxbuster":
			return "feroxbuster"
		default:
			return s
		}
	}

	t.Run("stealth filters low-noise and resolves display names", func(t *testing.T) {
		got := CapToolsWithEngine(display, "web", "stealth", eng, shortID)
		for _, n := range got {
			id := shortID(n)
			if _, ok := stealthToolIDs[id]; !ok {
				t.Fatalf("unexpected stealth tool %q (%s)", n, id)
			}
		}
		if len(got) > 4 {
			t.Fatalf("stealth cap: got %d tools %v", len(got), got)
		}
	})

	t.Run("comprehensive keeps high-effectiveness tools", func(t *testing.T) {
		got := CapToolsWithEngine(display, "web", "comprehensive", eng, shortID)
		if len(got) == 0 {
			t.Fatal("expected comprehensive filter to retain tools")
		}
		for _, n := range got {
			if eng.Score("web", shortID(n)) <= 0.7 {
				t.Fatalf("%q score too low for comprehensive", n)
			}
		}
	})

	t.Run("comprehensive empty filter falls back to CapTools", func(t *testing.T) {
		low := []string{"unknown-tool-a", "unknown-tool-b", "unknown-tool-c", "unknown-tool-d", "unknown-tool-e", "unknown-tool-f"}
		got := CapToolsWithEngine(low, "web", "comprehensive", eng, nil)
		want := CapTools(low, "comprehensive")
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("fallback: got %v want %v", got, want)
		}
	})

	t.Run("nil shortID uses identity", func(t *testing.T) {
		ids := []string{"httpx", "nuclei", "amass", "subfinder", "nmap"}
		got := CapToolsWithEngine(ids, "web", "stealth", eng, nil)
		if len(got) == 0 {
			t.Fatal("expected stealth tools")
		}
	})

	t.Run("default delegates to CapTools focused", func(t *testing.T) {
		got := CapToolsWithEngine(display, "web", "focused", eng, shortID)
		want := CapTools(display, "focused")
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v want %v", got, want)
		}
	})
}

func TestResolveNames(t *testing.T) {
	shortID := func(s string) string {
		if s == "Display Nuclei" {
			return "nuclei"
		}
		return s
	}
	original := []string{"Display Nuclei", "orphan-id"}
	tests := []struct {
		name     string
		shortIDs []string
		want     []string
	}{
		{
			name:     "maps short id to display name",
			shortIDs: []string{"nuclei"},
			want:     []string{"Display Nuclei"},
		},
		{
			name:     "unknown short id passes through",
			shortIDs: []string{"missing-tool"},
			want:     []string{"missing-tool"},
		},
		{
			name:     "deduplicates resolved names",
			shortIDs: []string{"nuclei", "nuclei"},
			want:     []string{"Display Nuclei"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveNames(tt.shortIDs, original, shortID)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("resolveNames() = %v, want %v", got, tt.want)
			}
		})
	}
}
