package intelligence

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

const probeTimeout = 5 * time.Second

// probeTarget performs lightweight HTTP/DNS heuristics (subset of HexStrike analyze_target).
func probeTarget(ctx context.Context, target string) (targetType string, technologies []string, cms string, confidence float64) {
	targetType = "unknown"
	confidence = 0.5
	technologies = []string{}

	if looksLikeCloudHost(target) {
		targetType = "cloud"
		technologies = append(technologies, "cloud")
		confidence = 0.65
	}
	if looksLikeBinary(target) {
		targetType = "binary"
		technologies = append(technologies, "binary")
		confidence = 0.7
		return
	}

	host := target
	if u, err := url.Parse(target); err == nil && u.Host != "" {
		host = u.Host
		if strings.Contains(strings.ToLower(u.Path), "/api") {
			targetType = "api"
		} else {
			targetType = "web"
		}
		technologies = append(technologies, "http")
		if c := detectCMSFromPath(u.Path); c != "" {
			cms = c
			technologies = append(technologies, c)
			confidence = 0.75
		}
	} else if ip := net.ParseIP(strings.Trim(host, "[]")); ip != nil {
		targetType = "ip"
		confidence = 0.8
	} else if strings.Count(host, ".") >= 1 && !strings.Contains(host, " ") {
		targetType = "web"
	}

	if targetType == "web" || targetType == "api" {
		if hdrs := httpProbe(ctx, normalizeURL(target)); len(hdrs) > 0 {
			technologies = mergeUnique(technologies, hdrs...)
			if cms == "" {
				cms = cmsFromHeaders(hdrs)
			}
			confidence = 0.8
		}
	}
	return
}

func normalizeURL(target string) string {
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		return target
	}
	return "https://" + target
}

func httpProbe(ctx context.Context, rawURL string) []string {
	ctx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, rawURL, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", "veil-engage/1.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// fallback GET for servers that reject HEAD
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
		if err != nil {
			return nil
		}
		req.Header.Set("User-Agent", "veil-engage/1.0")
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			return nil
		}
	}
	defer resp.Body.Close()
	var tech []string
	if s := resp.Header.Get("Server"); s != "" {
		tech = append(tech, "server:"+strings.ToLower(s))
	}
	if p := resp.Header.Get("X-Powered-By"); p != "" {
		tech = append(tech, "powered:"+strings.ToLower(p))
	}
	if ct := resp.Header.Get("Content-Type"); ct != "" {
		tech = append(tech, "content:"+strings.Split(ct, ";")[0])
	}
	return tech
}

func detectCMSFromPath(p string) string {
	low := strings.ToLower(p)
	switch {
	case strings.Contains(low, "wp-admin"), strings.Contains(low, "wp-content"):
		return "wordpress"
	case strings.Contains(low, "drupal"):
		return "drupal"
	case strings.Contains(low, "joomla"):
		return "joomla"
	}
	return ""
}

func cmsFromHeaders(tech []string) string {
	for _, t := range tech {
		low := strings.ToLower(t)
		if strings.Contains(low, "wordpress") {
			return "wordpress"
		}
		if strings.Contains(low, "drupal") {
			return "drupal"
		}
		if strings.Contains(low, "joomla") {
			return "joomla"
		}
		if strings.Contains(low, "php") {
			return "php"
		}
	}
	return ""
}

// technologiesDetected returns normalized technology labels from probe output.
func technologiesDetected(tech []string, cms string) []string {
	labels := []string{}
	if cms != "" {
		labels = append(labels, cms)
	}
	for _, t := range tech {
		low := strings.ToLower(t)
		switch {
		case strings.Contains(low, "nginx"):
			labels = append(labels, "nginx")
		case strings.Contains(low, "apache"):
			labels = append(labels, "apache")
		case strings.Contains(low, "php"):
			labels = append(labels, "php")
		case strings.Contains(low, "node") || strings.Contains(low, "express"):
			labels = append(labels, "nodejs")
		case strings.Contains(low, "java") || strings.Contains(low, "tomcat"):
			labels = append(labels, "java")
		case strings.Contains(low, "asp.net"):
			labels = append(labels, "dotnet")
		}
	}
	return mergeUnique(labels, nil...)
}

func looksLikeCloudHost(target string) bool {
	low := strings.ToLower(target)
	for _, hint := range []string{"amazonaws.com", "azure", "googleapis.com", "cloudfront.net", "s3."} {
		if strings.Contains(low, hint) {
			return true
		}
	}
	return false
}

func looksLikeBinary(target string) bool {
	ext := strings.ToLower(path.Ext(target))
	switch ext {
	case ".exe", ".elf", ".bin", ".so", ".dll", ".apk":
		return true
	}
	return false
}

func mergeUnique(base []string, add ...string) []string {
	seen := make(map[string]struct{}, len(base)+len(add))
	out := make([]string, 0, len(base)+len(add))
	for _, s := range append(base, add...) {
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func riskFromProfile(targetType string, techCount int) string {
	switch targetType {
	case "api", "web":
		if techCount > 3 {
			return "high"
		}
		return "medium"
	case "cloud":
		return "high"
	case "ip":
		return "medium"
	default:
		return "medium"
	}
}
