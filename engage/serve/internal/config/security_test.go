package config

import "testing"

func TestLoadSecurityForEnv_denyRawInProd(t *testing.T) {
	t.Setenv("ENGAGE_ALLOW_RAW_COMMAND", "1")
	t.Setenv("ENGAGE_DENY_RAW_COMMAND", "0")
	sec := LoadSecurityForEnv("prod")
	if sec.AllowRawCommand {
		t.Fatal("prod must deny raw commands")
	}
}

func TestLoadSecurityForEnv_denyRawFlag(t *testing.T) {
	t.Setenv("ENGAGE_ALLOW_RAW_COMMAND", "1")
	t.Setenv("ENGAGE_DENY_RAW_COMMAND", "1")
	sec := LoadSecurityForEnv("local")
	if sec.AllowRawCommand {
		t.Fatal("ENGAGE_DENY_RAW_COMMAND must disable raw")
	}
}
