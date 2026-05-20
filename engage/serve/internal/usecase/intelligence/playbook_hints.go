package intelligence

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/butbeautifulv/veil/engage/serve/internal/client/veilgraph"
)

var attackTechniqueRE = regexp.MustCompile(`\bT\d{4}(?:\.\d{3})?\b`)

const maxPlaybookHintTechniques = 5

// attachPlaybookHints adds veil-api playbook summaries for MITRE technique ids found in text blobs.
func attachPlaybookHints(ctx context.Context, veil veilgraph.Reader, out map[string]any, blobs ...string) {
	if veil == nil || !veil.Enabled() {
		return
	}
	ids := collectAttackTechniqueIDs(blobs...)
	if len(ids) == 0 {
		return
	}
	hints := make(map[string]json.RawMessage, len(ids))
	for _, tid := range ids {
		raw, err := veil.PlaybooksByTechnique(ctx, tid)
		if err != nil || !validGraphJSON(raw) {
			continue
		}
		hints[tid] = raw
	}
	if len(hints) > 0 {
		out["playbook_hints"] = hints
	}
}

func collectAttackTechniqueIDs(blobs ...string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, blob := range blobs {
		for _, m := range attackTechniqueRE.FindAllString(blob, -1) {
			tid := strings.ToUpper(m)
			if _, dup := seen[tid]; dup {
				continue
			}
			seen[tid] = struct{}{}
			out = append(out, tid)
			if len(out) >= maxPlaybookHintTechniques {
				return out
			}
		}
	}
	return out
}
