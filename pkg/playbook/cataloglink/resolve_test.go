package cataloglink

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func setRepoRoot(t *testing.T) {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
	if _, err := os.Stat(filepath.Join(root, "versions.env")); err != nil {
		t.Skip("repo root not found (versions.env)")
	}
	t.Setenv("VEIL_REPO_ROOT", root)
	ResetCatalogForTest()
}

func TestResolveMentions_table(t *testing.T) {
	setRepoRoot(t)

	tests := []struct {
		name     string
		mentions []string
		want     []string
	}{
		{
			name:     "short tokens skipped",
			mentions: []string{"dd", "dcfldd"},
			want:     nil,
		},
		{
			name:     "nmap alias",
			mentions: []string{"nmap"},
			want:     []string{"nmap_scan"},
		},
		{
			name:     "nuclei alias",
			mentions: []string{"nuclei"},
			want:     []string{"nuclei_scan"},
		},
		{
			name:     "dedup and order",
			mentions: []string{"nmap", "NMAP", "nmap"},
			want:     []string{"nmap_scan"},
		},
		{
			name:     "unknown tool",
			mentions: []string{"not-a-real-tool-xyz"},
			want:     nil,
		},
		{
			name:     "empty and whitespace",
			mentions: []string{"", "  ", "nmap"},
			want:     []string{"nmap_scan"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			setRepoRoot(t)
			got := ResolveMentions(tc.mentions)
			if len(got) != len(tc.want) {
				t.Fatalf("got %v want %v", got, tc.want)
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Fatalf("got %v want %v", got, tc.want)
				}
			}
		})
	}
}
