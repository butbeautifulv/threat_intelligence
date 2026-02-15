package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

var layerByPkg = map[string]string{
	"/internal/controller": "controller",
	"/internal/usecase":    "usecase",
	"/internal/domain":     "domain",
	"/internal/adapter":    "adapter",
}

var allowedDeps = map[string][]string{
	"controller": {"usecase"},
	"usecase":    {"domain", "adapter"},
	"domain":     {},
	"adapter":    {"domain"},
}

func main() {
	root := "./internal"
	var errs []string

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			errs = append(errs, fmt.Sprintf("parse error: %s: %v", path, err))
			return nil
		}

		pkgDir := filepath.Dir(path)
		layer := detectLayer(pkgDir)
		if layer == "" {
			return nil
		}

		for _, imp := range f.Imports {
			p := strings.Trim(imp.Path.Value, `"`)
			if !strings.Contains(p, "/internal/") {
				continue
			}
			targetLayer := detectLayerFromImport(p)
			if targetLayer == "" || targetLayer == layer {
				continue
			}
			if !isAllowed(layer, targetLayer) {
				errs = append(errs, fmt.Sprintf(
					"%s: %s (%s) -> %s (%s) is forbidden",
					path, pkgDir, layer, p, targetLayer,
				))
			}
		}
		return nil
	})

	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Println(e)
		}
		os.Exit(1)
	}
}

func detectLayer(dir string) string {
	for suffix, layer := range layerByPkg {
		if strings.Contains(dir, suffix) {
			return layer
		}
	}
	return ""
}

func detectLayerFromImport(imp string) string {
	for suffix, layer := range layerByPkg {
		if strings.Contains(imp, suffix) {
			return layer
		}
	}
	return ""
}

func isAllowed(from, to string) bool {
	for _, a := range allowedDeps[from] {
		if a == to {
			return true
		}
	}
	return false
}
