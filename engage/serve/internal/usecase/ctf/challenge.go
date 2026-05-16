package ctf

import (
	"fmt"
	"strings"
)

// Challenge describes a CTF challenge (HexStrike CTFChallenge parity).
type Challenge struct {
	Name        string   `json:"name"`
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Difficulty  string   `json:"difficulty"`
	Points      int      `json:"points"`
	Files       []string `json:"files,omitempty"`
	URL         string   `json:"url,omitempty"`
	Target      string   `json:"target,omitempty"`
	Hints       []string `json:"hints,omitempty"`
}

// Validate returns an error if required fields are missing.
func (c Challenge) Validate(requireName bool) error {
	if requireName && strings.TrimSpace(c.Name) == "" {
		return fmt.Errorf("challenge name is required")
	}
	cat := NormalizeCategory(c.Category)
	if cat == "" {
		c.Category = "misc"
	} else {
		c.Category = cat
	}
	if c.Difficulty == "" {
		c.Difficulty = "unknown"
	}
	if c.Points <= 0 {
		c.Points = 100
	}
	return nil
}

// NormalizeCategory lowercases and validates CTF category.
func NormalizeCategory(cat string) string {
	cat = strings.ToLower(strings.TrimSpace(cat))
	switch cat {
	case "web", "crypto", "pwn", "forensics", "rev", "misc", "osint":
		return cat
	default:
		return cat
	}
}

// TargetOrURL returns the best run target for tools.
func (c Challenge) TargetOrURL() string {
	if t := strings.TrimSpace(c.Target); t != "" {
		return t
	}
	if u := strings.TrimSpace(c.URL); u != "" {
		return u
	}
	return c.Name
}
