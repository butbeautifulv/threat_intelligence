package handle_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/butbeautifulv/threat_intelligence/pipeline/pipeline_worker/internal/handle"
	"github.com/butbeautifulv/threat_intelligence/scrape/contract/scrapev1"

	vulndomain "github.com/butbeautifulv/threat_intelligence/pipeline/internal/normalize/vuln"
)

func TestHandleVuln_NVDPage_CWEAndCPE(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "pkg", "nvdparse", "testdata", "nvd_page_min.json"))
	if err != nil {
		t.Fatal(err)
	}
	page := scrapev1.VulnNVDPage{RawJSON: string(raw)}
	payload, err := json.Marshal(page)
	if err != nil {
		t.Fatal(err)
	}
	env := &scrapev1.Envelope{Kind: scrapev1.KindVulnNVDPage, Payload: payload}
	out, err := handle.HandleVuln(context.Background(), env)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 2 {
		t.Fatalf("envelopes: got %d want 2", len(out))
	}
	var v vulndomain.Vulnerability
	if err := json.Unmarshal(out[0].Payload, &v); err != nil {
		t.Fatal(err)
	}
	if len(v.CWE) == 0 {
		t.Fatalf("expected CWE on first CVE, got %#v", v)
	}
	if len(v.CPEs) == 0 {
		t.Fatalf("expected CPEs on first CVE, got %#v", v)
	}
}
