package runner

import "github.com/butbeautifulv/veil/pkg/exec"

// catalogBinaryCandidates maps catalog binary names to PATH lookups (first match wins).
var catalogBinaryCandidates = map[string][]string{
	"burpsuite": {"burpsuite", "burpsuite_scan"},
}

// LookupBinary resolves a catalog binary to an executable on PATH.
func LookupBinary(name string) (string, error) {
	candidates := []string{name}
	if alts, ok := catalogBinaryCandidates[name]; ok {
		candidates = alts
	}
	var last error
	for _, c := range candidates {
		path, err := exec.LookupBinary(c)
		if err == nil {
			return path, nil
		}
		last = err
	}
	return "", last
}
