// Package execfetch documents optional subprocess fetch via pkg/exec.
//
// Default discovery harvest paths use HTTP (feeds.Client, githubraw) — no subprocess.
// Build with -tags discoveryexec to compile GitClone and related helpers for the
// discovery-fetcher container profile (see pkg/exec/README.md).
package execfetch
