package decision

import "strings"

// CapTools limits tool count by objective (quick/focused/stealth).
func CapTools(names []string, objective string) []string {
	switch strings.ToLower(strings.TrimSpace(objective)) {
	case "quick", "fast":
		if len(names) > 3 {
			return names[:3]
		}
	case "focused":
		if len(names) > 5 {
			return names[:5]
		}
	case "stealth":
		if len(names) > 4 {
			return names[:4]
		}
	}
	return names
}

// CapToolsWithEngine applies objective-specific caps using stealth/comprehensive filters.
// shortID maps a catalog display name to the effectiveness table tool id.
func CapToolsWithEngine(names []string, targetType, objective string, eng *DecisionEngine, shortID func(string) string) []string {
	if shortID == nil {
		shortID = func(s string) string { return s }
	}
	obj := strings.ToLower(strings.TrimSpace(objective))
	if obj == "stealth" {
		ids := make([]string, 0, len(names))
		for _, n := range names {
			ids = append(ids, shortID(n))
		}
		filtered := FilterStealthTools(ids)
		return resolveNames(filtered, names, shortID)
	}
	if obj == "comprehensive" && eng != nil {
		ids := make([]string, 0, len(names))
		for _, n := range names {
			ids = append(ids, shortID(n))
		}
		filtered := FilterComprehensiveTools(eng, targetType, ids)
		if len(filtered) > 0 {
			return resolveNames(filtered, names, shortID)
		}
	}
	return CapTools(names, objective)
}

func resolveNames(shortIDs []string, original []string, shortID func(string) string) []string {
	byShort := map[string]string{}
	for _, n := range original {
		byShort[shortID(n)] = n
	}
	out := make([]string, 0, len(shortIDs))
	seen := map[string]struct{}{}
	for _, id := range shortIDs {
		name, ok := byShort[id]
		if !ok {
			name = id
		}
		if _, dup := seen[name]; dup {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out
}
