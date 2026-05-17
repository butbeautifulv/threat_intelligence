package exec

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
	for _, want := range []string{"-sV", "-p", "80,443", "127.0.0.1"} {
		if !contains(args, want) {
			t.Fatalf("missing %q in %v", want, args)
		}
	}
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

func TestBuildArgs_goldenTemplates(t *testing.T) {
	cases := []struct {
		name       string
		template   []string
		target     string
		additional string
		params     map[string]string
		contains   []string
		absent     []string
	}{
		{
			name:     "rustscan_fast_scan",
			template: []string{"-a", "{target}", "-p", "{ports}", "{additional_args}"},
			target:   "10.0.0.1",
			params:   map[string]string{"ports": "80,443"},
			contains: []string{"-a", "10.0.0.1", "-p", "80,443"},
		},
		{
			name:     "gobuster_scan",
			template: []string{"{mode}", "-u", "{target}", "-w", "{wordlist}", "{additional_args}"},
			target:   "http://127.0.0.1",
			params:   map[string]string{"mode": "dir", "wordlist": "/tmp/w.txt"},
			contains: []string{"dir", "-u", "http://127.0.0.1", "-w", "/tmp/w.txt"},
		},
		{
			name:     "ffuf_scan",
			template: []string{"-u", "{target}", "-w", "{wordlist}", "{additional_args}"},
			target:   "http://test/",
			params:   map[string]string{"wordlist": "/wl.txt"},
			contains: []string{"-u", "http://test/", "-w", "/wl.txt"},
		},
		{
			name:     "sqlmap_scan",
			template: []string{"-u", "{target}", "--data", "{data}", "{additional_args}"},
			target:   "http://vuln/",
			params:   map[string]string{"data": "id=1"},
			contains: []string{"-u", "http://vuln/", "--data", "id=1"},
		},
		{
			name:     "nikto_scan",
			template: []string{"-h", "{target}", "{additional_args}"},
			target:   "example.com",
			params:   nil,
			contains: []string{"-h", "example.com"},
		},
		{
			name:     "nuclei_scan",
			template: []string{"-u", "{target}", "-t", "{templates}", "{additional_args}"},
			target:   "http://x/",
			params:   map[string]string{"templates": "cves/"},
			contains: []string{"-u", "http://x/", "-t", "cves/"},
		},
		{
			name:     "masscan_high_speed",
			template: []string{"{target}", "-p", "{ports}", "--rate", "{rate}", "{additional_args}"},
			target:   "10.0.0.0/24",
			params:   map[string]string{"ports": "1-1024", "rate": "500"},
			contains: []string{"10.0.0.0/24", "-p", "1-1024", "--rate", "500"},
		},
		{
			name:     "hydra_attack",
			template: []string{"-l", "{username}", "-P", "{password_file}", "{target}", "-s", "{service}", "{additional_args}"},
			target:   "10.0.0.5",
			params:   map[string]string{"username": "admin", "password_file": "/tmp/pw.txt", "service": "ssh"},
			contains: []string{"-l", "admin", "-P", "/tmp/pw.txt", "10.0.0.5", "-s", "ssh"},
		},
		{
			name:       "dalfox_xss_scan",
			template:   []string{"url", "{target}", "{additional_args}"},
			target:     "http://vuln.test/",
			additional: "--silence",
			contains:   []string{"url", "http://vuln.test/", "--silence"},
		},
		{
			name:     "naabu_port_scan",
			template: []string{"-host", "{target}", "-p", "{ports}", "{additional_args}"},
			target:   "10.0.0.1",
			params:   map[string]string{"ports": "80,443"},
			contains: []string{"-host", "10.0.0.1", "-p", "80,443"},
		},
		{
			name:     "subfinder_scan",
			template: []string{"-d", "{target}", "{additional_args}"},
			target:   "example.com",
			contains: []string{"-d", "example.com"},
		},
		{
			name:     "dirsearch_scan",
			template: []string{"-u", "{target}", "-e", "{extensions}", "{additional_args}"},
			target:   "http://127.0.0.1",
			params:   map[string]string{"extensions": "php,html"},
			contains: []string{"-u", "http://127.0.0.1", "-e", "php,html"},
		},
		{
			name:     "testssl_scan",
			template: []string{"{target}", "{additional_args}"},
			target:   "example.com:443",
			contains: []string{"example.com:443"},
		},
		{
			name:     "gau_discovery",
			template: []string{"--subs", "{target}", "{additional_args}"},
			target:   "example.com",
			contains: []string{"--subs", "example.com"},
		},
		{
			name:     "arjun_scan",
			template: []string{"-u", "{target}", "-m", "{method}", "{additional_args}"},
			target:   "http://api.test/",
			params:   map[string]string{"method": "GET"},
			contains: []string{"-u", "http://api.test/", "-m", "GET"},
		},
		{
			name:     "trivy_scan",
			template: []string{"image", "{target}", "{additional_args}"},
			target:   "alpine:latest",
			contains: []string{"image", "alpine:latest"},
		},
		{
			name:     "amass_scan",
			template: []string{"{mode}", "-d", "{target}", "{additional_args}"},
			target:   "corp.example",
			params:   map[string]string{"mode": "enum"},
			contains: []string{"enum", "-d", "corp.example"},
		},
		{
			name:     "fierce_scan",
			template: []string{"--domain", "{target}", "{additional_args}"},
			target:   "example.org",
			contains: []string{"--domain", "example.org"},
		},
		{
			name:       "katana_crawl",
			template:   []string{"-u", "{target}", "{additional_args}"},
			target:     "https://example.com",
			additional: "-d 2",
			contains:   []string{"-u", "https://example.com", "-d", "2"},
		},
		{
			name:       "wpscan_analyze",
			template:   []string{"--url", "{target}", "{additional_args}"},
			target:     "https://wp.example.com",
			additional: "--enumerate p",
			contains:   []string{"--url", "https://wp.example.com", "--enumerate", "p"},
		},
		{
			name:       "feroxbuster_scan",
			template:   []string{"-u", "{target}", "-w", "{wordlist}", "{additional_args}"},
			target:     "http://127.0.0.1/",
			additional: "-q",
			params:     map[string]string{"wordlist": "/wl.txt"},
			contains:   []string{"-u", "http://127.0.0.1/", "-w", "/wl.txt", "-q"},
		},
		{
			name:       "httpx_probe",
			template:   []string{"-u", "{target}", "{additional_args}"},
			target:     "https://example.com",
			additional: "-silent",
			contains:   []string{"-u", "https://example.com", "-silent"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			args := BuildArgs(tc.template, tc.target, tc.additional, tc.params)
			for _, want := range tc.contains {
				if !contains(args, want) {
					t.Fatalf("missing %q in %v", want, args)
				}
			}
			for _, bad := range tc.absent {
				if contains(args, bad) {
					t.Fatalf("unexpected %q in %v", bad, args)
				}
			}
		})
	}
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
