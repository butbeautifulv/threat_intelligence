// Command executable-matrix probes every catalog tool through tooldispatch (CI matrix gate).
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/tooldispatch"
)

func main() {
	root := flag.String("root", "", "repository root (default: auto from cwd)")
	catalog := flag.String("catalog", "", "path to tools.yaml")
	flag.Parse()

	repoRoot := *root
	if repoRoot == "" {
		repoRoot = findRepoRoot()
	}
	catalogPath := *catalog
	if catalogPath == "" {
		catalogPath = filepath.Join(repoRoot, "engage", "serve", "catalog", "tools.yaml")
	}

	names, err := catalogToolNames(catalogPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "catalog: %v\n", err)
		os.Exit(2)
	}

	d, err := tooldispatch.NewMatrixDispatcher(catalogPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dispatcher: %v\n", err)
		os.Exit(2)
	}

	results := tooldispatch.RunMatrix(context.Background(), d, names)
	pass := 0
	for _, r := range results {
		if r.Pass {
			pass++
		}
		enc := json.NewEncoder(os.Stdout)
		_ = enc.Encode(r)
	}
	fmt.Fprintf(os.Stderr, "MATRIX_TOTAL=%d\nMATRIX_PASS=%d\n", len(names), pass)
	if pass < len(names) {
		os.Exit(1)
	}
}

func findRepoRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "engage", "serve", "catalog", "tools.yaml")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return wd
}

var catalogNameRE = regexp.MustCompile(`(?m)^  - name: (\S+)`)

func catalogToolNames(catalogPath string) ([]string, error) {
	b, err := os.ReadFile(catalogPath)
	if err != nil {
		return nil, err
	}
	names := catalogNameRE.FindAllStringSubmatch(string(b), -1)
	if len(names) == 0 {
		return nil, fmt.Errorf("no tools in %s", catalogPath)
	}
	out := make([]string, 0, len(names))
	for _, m := range names {
		out = append(out, m[1])
	}
	sort.Strings(out)
	return out, nil
}
