// Package main is deprecated; browser crawl lives under discovery/cmd/browser-agent.
package main

import "os"

func main() {
	os.Stderr.WriteString("engage browser-agent moved to discovery/cmd/browser-agent; use discovery-browser service (DISCOVERY_BROWSER_URL)\n")
	os.Exit(1)
}
