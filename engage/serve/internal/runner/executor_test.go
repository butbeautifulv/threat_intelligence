package runner

import "testing"

func TestBuildArgs_nmapWithPorts(t *testing.T) {
	args := BuildArgs(
		[]string{"{scan_type}", "-p", "{ports}", "{additional_args}", "{target}"},
		"127.0.0.1",
		"",
		map[string]string{"scan_type": "-sV", "ports": "80,443"},
	)
	if len(args) < 3 {
		t.Fatalf("args too short: %v", args)
	}
	joined := joinArgs(args)
	for _, want := range []string{"-sV", "-p", "80,443", "127.0.0.1"} {
		if !contains(args, want) {
			t.Fatalf("missing %q in %v", want, args)
		}
	}
	_ = joined
}

func TestBuildArgs_skipEmptyPorts(t *testing.T) {
	args := BuildArgs(
		[]string{"{scan_type}", "-p", "{ports}", "{target}"},
		"10.0.0.1",
		"",
		map[string]string{"scan_type": "-sV", "ports": ""},
	)
	for _, a := range args {
		if a == "-p" {
			t.Fatalf("should skip -p when ports empty: %v", args)
		}
	}
}

func joinArgs(a []string) string {
	s := ""
	for _, x := range a {
		s += x + " "
	}
	return s
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
