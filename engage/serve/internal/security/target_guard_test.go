package security

import "testing"

func TestCheckTarget_blocksMetadata(t *testing.T) {
	blocked, _ := CheckTarget("http://169.254.169.254/latest/meta-data/")
	if !blocked {
		t.Fatal("expected metadata IP blocked")
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
