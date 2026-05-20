package report

import (
	"encoding/json"
	"testing"

	engageevents "github.com/butbeautifulv/veil/pkg/engage/events"
)

func TestFinding_ToFindingEvent(t *testing.T) {
	f := Finding{
		Title:       "Reflected XSS",
		Severity:    SeverityHigh,
		Description: "param q reflects input",
		Target:      "https://example.com",
		Tool:        "nuclei",
		Evidence:    "screenshot.png",
	}
	ev := f.ToFindingEvent()
	want := engageevents.FindingEvent{
		Tool:        "nuclei",
		Target:      "https://example.com",
		Title:       "Reflected XSS",
		Severity:    "high",
		Description: "param q reflects input",
	}
	if ev != want {
		t.Fatalf("got %+v, want %+v", ev, want)
	}
	b, err := json.Marshal(ev)
	if err != nil {
		t.Fatal(err)
	}
	var round engageevents.FindingEvent
	if err := json.Unmarshal(b, &round); err != nil {
		t.Fatal(err)
	}
	if round != want {
		t.Fatalf("round-trip %+v", round)
	}
}

func TestSeverityConstants(t *testing.T) {
	want := []Severity{
		SeverityInfo,
		SeverityLow,
		SeverityMedium,
		SeverityHigh,
		SeverityCritical,
	}
	for _, s := range want {
		if s == "" {
			t.Fatal("severity constant must not be empty")
		}
	}
}
