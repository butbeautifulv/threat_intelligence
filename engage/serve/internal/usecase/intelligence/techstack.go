package intelligence

import (
	"context"
	"net/http"
	"strings"
)

// Technology identifies a detected stack component (HexStrike TechnologyStack parity).
type Technology string

const (
	TechApache    Technology = "apache"
	TechNginx     Technology = "nginx"
	TechIIS       Technology = "iis"
	TechNodeJS    Technology = "nodejs"
	TechPHP       Technology = "php"
	TechPython    Technology = "python"
	TechJava      Technology = "java"
	TechDotNet    Technology = "dotnet"
	TechWordPress Technology = "wordpress"
	TechDrupal    Technology = "drupal"
	TechJoomla    Technology = "joomla"
	TechReact     Technology = "react"
	TechAngular   Technology = "angular"
	TechVue       Technology = "vue"
	TechUnknown   Technology = "unknown"
)

// AllTechnologies returns the canonical technology enum values.
func AllTechnologies() []Technology {
	return []Technology{
		TechApache, TechNginx, TechIIS, TechNodeJS, TechPHP, TechPython, TechJava, TechDotNet,
		TechWordPress, TechDrupal, TechJoomla, TechReact, TechAngular, TechVue, TechUnknown,
	}
}

func (t Technology) String() string { return string(t) }

// DetectTechnologies infers stack from target string, HTTP headers, and optional body snippet.
func DetectTechnologies(ctx context.Context, target string, headers http.Header, body string) []Technology {
	seen := map[Technology]struct{}{}
	add := func(ts ...Technology) {
		for _, t := range ts {
			if t != "" {
				seen[t] = struct{}{}
			}
		}
	}

	low := strings.ToLower(target + " " + body)
	if strings.Contains(low, "wordpress") || strings.Contains(low, "wp-content") || strings.Contains(low, "wp-admin") {
		add(TechWordPress, TechPHP)
	}
	if strings.Contains(low, "drupal") {
		add(TechDrupal, TechPHP)
	}
	if strings.Contains(low, "joomla") {
		add(TechJoomla, TechPHP)
	}
	if strings.Contains(low, ".php") {
		add(TechPHP)
	}
	if strings.Contains(low, ".asp") || strings.Contains(low, "asp.net") {
		add(TechDotNet)
	}
	if strings.Contains(low, "react") || strings.Contains(low, "__react") {
		add(TechReact)
	}
	if strings.Contains(low, "angular") || strings.Contains(low, "ng-version") {
		add(TechAngular)
	}
	if strings.Contains(low, "vue") || strings.Contains(low, "__vue") {
		add(TechVue)
	}

	if headers != nil {
		if s := strings.ToLower(headers.Get("Server")); s != "" {
			switch {
			case strings.Contains(s, "nginx"):
				add(TechNginx)
			case strings.Contains(s, "apache"):
				add(TechApache)
			case strings.Contains(s, "iis") || strings.Contains(s, "microsoft"):
				add(TechIIS)
			}
		}
		if p := strings.ToLower(headers.Get("X-Powered-By")); p != "" {
			switch {
			case strings.Contains(p, "php"):
				add(TechPHP)
			case strings.Contains(p, "express") || strings.Contains(p, "node"):
				add(TechNodeJS)
			case strings.Contains(p, "asp.net"):
				add(TechDotNet)
			case strings.Contains(p, "python") || strings.Contains(p, "django") || strings.Contains(p, "flask"):
				add(TechPython)
			}
		}
	}

	if len(seen) == 0 {
		add(TechUnknown)
	}

	out := make([]Technology, 0, len(seen))
	for t := range seen {
		if t == TechUnknown && len(seen) > 1 {
			continue
		}
		out = append(out, t)
	}
	if len(out) == 0 {
		return []Technology{TechUnknown}
	}
	return out
}

// TechnologiesToStrings converts enum slice to API string labels.
func TechnologiesToStrings(tech []Technology) []string {
	out := make([]string, 0, len(tech))
	for _, t := range tech {
		out = append(out, string(t))
	}
	return out
}

// techStackBoost maps detected technologies to tool score boosts.
func techStackBoost(tech []Technology) map[string]float64 {
	boost := map[string]float64{}
	for _, t := range tech {
		switch t {
		case TechWordPress:
			boost["wpscan"] = 0.25
			boost["nuclei"] = 0.05
		case TechPHP:
			boost["nikto"] = 0.15
			boost["sqlmap"] = 0.12
		case TechDrupal, TechJoomla:
			boost["nuclei"] = 0.1
			boost["nikto"] = 0.1
		case TechNodeJS:
			boost["nuclei"] = 0.08
			boost["ffuf"] = 0.05
		case TechJava:
			boost["nuclei"] = 0.08
		case TechDotNet:
			boost["nuclei"] = 0.06
		case TechNginx, TechApache:
			boost["nikto"] = 0.05
		}
	}
	if len(boost) == 0 {
		return nil
	}
	return boost
}
