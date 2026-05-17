package harvest

import (
	"testing"
)

func TestEnvelopeValidate(t *testing.T) {
	env, err := NewEnvelope(SourceDS, KindDSSigmaRaw, "ds:sigma:rules/x.yml", DSSigmaRaw{Path: "rules/x.yml", RawYAML: "id: x"})
	if err != nil {
		t.Fatal(err)
	}
	if err := env.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestEnvelopeBrowserInspect(t *testing.T) {
	env, err := NewEnvelope(SourceBrowser, KindBrowserInspectRaw, BrowserContentKey("https://example.com"), BrowserInspectRaw{
		URL: "https://example.com", RawJSON: `{"success":true}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := env.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestEnvelopeValidateEmptyPayload(t *testing.T) {
	env := &Envelope{SchemaVersion: 1, Source: SourceDS, Kind: KindDSSigmaRaw, ContentKey: "k"}
	if err := env.Validate(); err == nil {
		t.Fatal("expected error")
	}
}
