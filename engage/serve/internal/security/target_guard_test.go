package security

import (
	"strings"
	"testing"
)

func TestCheckTarget_blocksMetadata(t *testing.T) {
	blocked, reason := CheckTarget("http://169.254.169.254/latest/meta-data/")
	if !blocked {
		t.Fatal("expected metadata IP blocked")
	}
	if reason == "" {
		t.Fatal("expected non-empty reason")
	}
}

func TestCheckTarget_blocksLoopback(t *testing.T) {
	blocked, _ := CheckTarget("127.0.0.1")
	if !blocked {
		t.Fatal("expected loopback blocked")
	}
}

func TestCheckTarget_allowsPublicHost(t *testing.T) {
	blocked, _ := CheckTarget("https://example.com")
	if blocked {
		t.Fatal("public host should not be blocked")
	}
}

func TestCheckTarget_table(t *testing.T) {
	tests := []struct {
		target  string
		blocked bool
		contain string
	}{
		{"http://169.254.169.254/", true, "metadata"},
		{"169.254.169.254", true, "metadata"},
		{"127.0.0.1:8080", true, "loopback"},
		{"http://127.0.0.1/admin", true, "loopback"},
		{"localhost", true, "localhost"},
		{"http://localhost/", true, "localhost"},
		{"10.0.0.1", true, "RFC1918"},
		{"192.168.1.50", true, "RFC1918"},
		{"172.16.0.1", true, "RFC1918"},
		{"https://example.com", false, ""},
		{"8.8.8.8", false, ""},
		{"", false, ""},
	}
	for _, tc := range tests {
		t.Run(tc.target, func(t *testing.T) {
			blocked, reason := CheckTarget(tc.target)
			if blocked != tc.blocked {
				t.Fatalf("blocked=%v reason=%q", blocked, reason)
			}
			if tc.contain != "" && !strings.Contains(strings.ToLower(reason), strings.ToLower(tc.contain)) {
				t.Fatalf("reason %q should mention %q", reason, tc.contain)
			}
		})
	}
}

func TestEnforceTarget_denylistAlwaysBlocksMetadata(t *testing.T) {
	for _, mode := range []TargetGuardMode{TargetGuardOff, TargetGuardWarn, TargetGuardBlock} {
		t.Run(string(mode), func(t *testing.T) {
			blocked, reason := EnforceTarget("http://169.254.169.254/latest/meta-data/", mode)
			if !blocked {
				t.Fatalf("mode %q: expected metadata blocked", mode)
			}
			if !strings.Contains(strings.ToLower(reason), "metadata") {
				t.Fatalf("reason %q", reason)
			}
		})
	}
}

func TestEnforceTarget_privateOnlyWhenBlock(t *testing.T) {
	blocked, _ := EnforceTarget("10.0.0.1", TargetGuardOff)
	if blocked {
		t.Fatal("RFC1918 should be allowed when guard is off")
	}
	blocked, _ = EnforceTarget("10.0.0.1", TargetGuardBlock)
	if !blocked {
		t.Fatal("RFC1918 should be blocked when guard is block")
	}
}

func TestParseTargetGuardMode_prodDefault(t *testing.T) {
	m := ParseTargetGuardMode(func(k string) string {
		if k == "ENGAGE_ENV" {
			return "prod"
		}
		return ""
	})
	if m != TargetGuardBlock {
		t.Fatalf("got %q", m)
	}
}

func TestParseTargetGuardMode_table(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want TargetGuardMode
	}{
		{"explicit_off", map[string]string{"ENGAGE_TARGET_GUARD": "off"}, TargetGuardOff},
		{"explicit_warn", map[string]string{"ENGAGE_TARGET_GUARD": "warn"}, TargetGuardWarn},
		{"explicit_block", map[string]string{"ENGAGE_TARGET_GUARD": "block"}, TargetGuardBlock},
		{"alias_true", map[string]string{"ENGAGE_TARGET_GUARD": "true"}, TargetGuardBlock},
		{"non_prod_unset", map[string]string{"ENGAGE_ENV": "dev"}, TargetGuardOff},
		{"prod_implicit_block", map[string]string{"ENGAGE_ENV": "prod"}, TargetGuardBlock},
		{"prod_explicit_off", map[string]string{"ENGAGE_ENV": "prod", "ENGAGE_TARGET_GUARD": "off"}, TargetGuardOff},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			getenv := func(k string) string { return tc.env[k] }
			if got := ParseTargetGuardMode(getenv); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}
