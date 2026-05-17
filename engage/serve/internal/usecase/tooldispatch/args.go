package tooldispatch

import (
	"fmt"
	"strings"
)

func argTarget(args map[string]any) string {
	if args == nil {
		return ""
	}
	for _, k := range []string{"target", "base_url", "url", "domain", "host"} {
		if v, ok := args[k]; ok {
			if s := strings.TrimSpace(fmt.Sprint(v)); s != "" {
				return s
			}
		}
	}
	return ""
}

func argString(args map[string]any, key, def string) string {
	if args == nil {
		return def
	}
	if v, ok := args[key]; ok {
		if s := strings.TrimSpace(fmt.Sprint(v)); s != "" {
			return s
		}
	}
	return def
}

func argInt(args map[string]any, key string, def int) int {
	if args == nil {
		return def
	}
	switch v := args[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	default:
		return def
	}
}

func argBool(args map[string]any, key string) bool {
	if args == nil {
		return false
	}
	switch v := args[key].(type) {
	case bool:
		return v
	case string:
		return strings.EqualFold(v, "true") || v == "1"
	}
	return false
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
