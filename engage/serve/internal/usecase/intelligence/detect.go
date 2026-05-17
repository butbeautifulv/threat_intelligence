package intelligence

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

const maxBodySample = 64 << 10 // 64KB

const probeTimeout = 5 * time.Second

func init() {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.ForceAttemptHTTP2 = false
	tr.TLSHandshakeTimeout = probeTimeout
	tr.ResponseHeaderTimeout = probeTimeout
	probeHTTPClient = &http.Client{Timeout: probeTimeout, Transport: tr}
}

var probeHTTPClient *http.Client

// probeTarget performs lightweight HTTP/DNS heuristics (subset of HexStrike analyze_target).
func probeTarget(ctx context.Context, target string) (targetType string, technologies []string, cms string, confidence float64, hdr http.Header, body string) {
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
		return targetType, technologies, cms, confidence, hdr, body
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
		hdr, body, hdrLabels := httpProbe(ctx, normalizeURL(target))
		if len(hdrLabels) > 0 {
			technologies = mergeUnique(technologies, hdrLabels...)
			if cms == "" {
				cms = cmsFromHeaders(hdrLabels)
			}
			confidence = 0.8
		}
		if len(hdr) > 0 || body != "" {
			for _, t := range MatchHeaderSignatures(hdr) {
				technologies = mergeUnique(technologies, string(t))
			}
			for _, t := range MatchContentSignatures(body) {
				technologies = mergeUnique(technologies, string(t))
				if cms == "" && (t == TechWordPress || t == TechDrupal || t == TechJoomla) {
					cms = string(t)
				}
			}
		}
	}
	return targetType, technologies, cms, confidence, hdr, body
}

func normalizeURL(target string) string {
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		return target
	}
	return "https://" + target
}

// httpProbe fetches headers and a body sample for technology detection.
func httpProbe(ctx context.Context, rawURL string) (http.Header, string, []string) {
	ctx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, "", nil
	}
	req.Header.Set("User-Agent", "veil-engage/1.0")
	resp, err := probeHTTPClient.Do(req)
	if err != nil {
		return nil, "", nil
	}
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, maxBodySample))
	body := string(bodyBytes)
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
	return resp.Header.Clone(), body, tech
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

// resolveTargetIPs resolves hostname to IPs (HexStrike _resolve_domain subset).
func resolveTargetIPs(ctx context.Context, target string) []string {
	host := target
	if u, err := url.Parse(target); err == nil && u.Hostname() != "" {
		host = u.Hostname()
	}
	if host == "" || net.ParseIP(host) != nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil
	}
	out := make([]string, 0, len(addrs))
	for _, a := range addrs {
		if ip := a.IP.String(); ip != "" {
			out = append(out, ip)
		}
	}
	return mergeUnique(out, nil...)
}
