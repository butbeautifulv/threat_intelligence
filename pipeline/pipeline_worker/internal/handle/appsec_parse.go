package handle

import (
	"encoding/json"
	"strings"

	"github.com/butbeautifulv/threat_intelligence/pipeline/contract/ingestv1"
	"github.com/butbeautifulv/threat_intelligence/scrape/contract/scrapev1"

	"github.com/butbeautifulv/threat_intelligence/pipeline/internal/normalize/appsec"

	"gopkg.in/yaml.v3"
)

func coderulesSemgrepRaw(payload json.RawMessage) ([]*ingestv1.Envelope, error) {
	var raw scrapev1.CoderulesSemgrepRaw
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, err
	}
	var root map[string]any
	if err := yaml.Unmarshal([]byte(raw.RawYAML), &root); err != nil {
		return nil, err
	}
	ruleID, title := appsecparse.SemgrepMeta(root, raw.Path)
	lang := strings.Split(raw.Path, "/")[0]
	cwes := appsecparse.SemgrepCWES(root)
	pl := ingestv1.CoderulesSemgrepPayload{
		Path: raw.Path, Language: lang, RuleID: ruleID, Title: title, RawYAML: raw.RawYAML, CWEs: cwes,
	}
	out, err := ingestv1.NewEnvelope(ingestv1.SourceCoderules, ingestv1.KindCoderulesSemgrep, ingestv1.CoderulesSemgrepIdempotencyKey(raw.Path), pl)
	return []*ingestv1.Envelope{out}, err
}

func coderulesCodeQLRaw(payload json.RawMessage) ([]*ingestv1.Envelope, error) {
	var raw scrapev1.CoderulesCodeQLRaw
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, err
	}
	cwes := appsecparse.CodeQLCWES(raw.Body)
	pl := ingestv1.CoderulesCodeQLPayload{
		Path: raw.Path, Name: raw.Path, Lang: "javascript", Body: raw.Body, CWEs: cwes,
	}
	out, err := ingestv1.NewEnvelope(ingestv1.SourceCoderules, ingestv1.KindCoderulesCodeQL, ingestv1.CoderulesCodeQLIdempotencyKey(raw.Path), pl)
	return []*ingestv1.Envelope{out}, err
}

func nucleiTemplateRaw(payload json.RawMessage) ([]*ingestv1.Envelope, error) {
	var raw scrapev1.NucleiTemplateRaw
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, err
	}
	p, err := appsecparse.ParseNucleiYAML([]byte(raw.RawYAML))
	if err != nil || p.TemplateID == "" {
		return nil, err
	}
	pl := ingestv1.NucleiTemplatePayload{
		Path: raw.Path, TemplateID: p.TemplateID, Name: p.Name, Severity: p.Severity,
		TagsJSON: p.TagsJSON, CVE: p.CVE, CWE: p.CWE, RawYAML: raw.RawYAML,
	}
	out, err := ingestv1.NewEnvelope(ingestv1.SourceNuclei, ingestv1.KindNucleiTemplate, ingestv1.NucleiTemplateIdempotencyKey(raw.Path), pl)
	return []*ingestv1.Envelope{out}, err
}
