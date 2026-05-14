// Package ingestv1 defines a versioned JSON envelope for scraper → queue → worker ingest.
package ingestv1

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const CurrentSchemaVersion = 1

// Well-known source identifiers (scrapers).
const (
	SourceSBOM      = "sbom"
	SourceCoderules = "coderules"
	SourceNuclei    = "nuclei"
	SourceTI        = "ti"
)

// Event kinds per source (extend as needed).
const (
	KindSBOMOSVRecord    = "sbom_osv_record"
	KindSBOMGHSADocument = "sbom_ghsa_document"
	KindCoderulesCWERow  = "coderules_cwe_row"
	KindCoderulesSemgrep = "coderules_semgrep_yaml"
	KindCoderulesCodeQL  = "coderules_codeql_ql"
	KindNucleiTemplate   = "nuclei_template_yaml"

	// TI (threat intel feeds / JSONL) — payload is JSON matching ti/internal/domain and ti/internal/ingest shapes.
	KindTIIoC                   = "ti_ioc"
	KindTIKEVVulnerability      = "ti_kev_vulnerability"
	KindTIReport                = "ti_report"
	KindTICampaign              = "ti_campaign"
	KindTICluster               = "ti_cluster"
	KindTIActor                 = "ti_actor"
	KindTILinkCampaignIOC       = "ti_link_campaign_ioc"
	KindTILinkClusterCampaign    = "ti_link_cluster_campaign"
	KindTILinkCampaignActor      = "ti_link_campaign_actor"
	KindTILinkReportMentionsIOC  = "ti_link_report_mentions_ioc"
	KindTIJSONLRecord            = "ti_jsonl_record"
)

// Envelope is the on-wire JSON for JetStream / HTTP bridges.
type Envelope struct {
	SchemaVersion  int             `json:"schema_version"`
	Source         string          `json:"source"`
	Kind           string          `json:"kind"`
	IdempotencyKey string          `json:"idempotency_key"`
	Payload        json.RawMessage `json:"payload"`
}

// Validate checks required fields for schema v1.
func (e *Envelope) Validate() error {
	if e.SchemaVersion != CurrentSchemaVersion {
		return fmt.Errorf("ingestv1: unsupported schema_version %d (want %d)", e.SchemaVersion, CurrentSchemaVersion)
	}
	if strings.TrimSpace(e.Source) == "" {
		return errors.New("ingestv1: empty source")
	}
	if strings.TrimSpace(e.Kind) == "" {
		return errors.New("ingestv1: empty kind")
	}
	if strings.TrimSpace(e.IdempotencyKey) == "" {
		return errors.New("ingestv1: empty idempotency_key")
	}
	if len(e.Payload) == 0 || string(e.Payload) == "null" {
		return errors.New("ingestv1: empty payload")
	}
	return nil
}

// SBOMOSVPayload is the payload for KindSBOMOSVRecord.
type SBOMOSVPayload struct {
	OSVID    string           `json:"osv_id"`
	CVE      string           `json:"cve"`
	Affected []map[string]any `json:"affected"`
}

// SBOMGHSAPathPayload carries GHSA JSON plus stable path for idempotency.
type SBOMGHSAPathPayload struct {
	Path string         `json:"path"`
	Doc  map[string]any `json:"doc"`
}

// CoderulesCWEPayload is one CWE catalog row.
type CoderulesCWEPayload struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

// CoderulesSemgrepPayload is raw YAML body + metadata for storage.
type CoderulesSemgrepPayload struct {
	Path     string   `json:"path"`
	Language string   `json:"language"`
	RuleID   string   `json:"rule_id"`
	Title    string   `json:"title"`
	RawYAML  string   `json:"raw_yaml"`
	CWEs     []string `json:"cwes"`
}

// CoderulesCodeQLPayload is a CodeQL query file snapshot.
type CoderulesCodeQLPayload struct {
	Path string   `json:"path"`
	Name string   `json:"name"`
	Lang string   `json:"lang"`
	Body string   `json:"body"`
	CWEs []string `json:"cwes"`
}

// NucleiTemplatePayload is parsed template fields + raw YAML.
type NucleiTemplatePayload struct {
	Path       string `json:"path"`
	TemplateID string `json:"template_id"`
	Name       string `json:"name"`
	Severity   string `json:"severity"`
	TagsJSON   string `json:"tags_json"`
	CVE        string `json:"cve"`
	CWE        string `json:"cwe"`
	RawYAML    string `json:"raw_yaml"`
}

// TIKEVVulnPayload is one CISA KEV row (same fields as ti feeds runner).
type TIKEVVulnPayload struct {
	CVEID         string `json:"cve_id"`
	VendorProject string `json:"vendor_project"`
	Product       string `json:"product"`
	ShortDesc     string `json:"short_desc"`
	DateAdded     string `json:"date_added"`
}

// TILinkCampaignIOCPayload links Campaign → IOC after both upserts.
type TILinkCampaignIOCPayload struct {
	CampaignID string          `json:"campaign_id"`
	IOC        json.RawMessage `json:"ioc"`
}

// TILinkClusterCampaignPayload links Cluster → Campaign.
type TILinkClusterCampaignPayload struct {
	ClusterID  string `json:"cluster_id"`
	CampaignID string `json:"campaign_id"`
}

// TILinkCampaignActorPayload links Campaign → Actor (actor identified by name).
type TILinkCampaignActorPayload struct {
	CampaignID string `json:"campaign_id"`
	ActorName  string `json:"actor_name"`
}

// TILinkReportMentionsIOCPayload links Report → IOC.
type TILinkReportMentionsIOCPayload struct {
	ReportID string          `json:"report_id"`
	IOC      json.RawMessage `json:"ioc"`
}

// TIJSONLRecordPayload is one raw JSONL line (ti/internal/ingest.Envelope JSON).
type TIJSONLRecordPayload struct {
	Line json.RawMessage `json:"line"`
}

// NewEnvelope builds a v1 envelope with JSON-marshaled payload.
func NewEnvelope(source, kind, idempotencyKey string, payload any) (*Envelope, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return &Envelope{
		SchemaVersion:  CurrentSchemaVersion,
		Source:         source,
		Kind:           kind,
		IdempotencyKey: idempotencyKey,
		Payload:        raw,
	}, nil
}

// SBOMOSVIdempotencyKey builds a stable key per CVE + package row (ecosystem:name).
func SBOMOSVIdempotencyKey(cve, ecosystem, pkgName string) string {
	cve = strings.TrimSpace(strings.ToUpper(cve))
	return fmt.Sprintf("sbom:osv:%s:%s:%s", cve, strings.ToLower(strings.TrimSpace(ecosystem)), strings.TrimSpace(pkgName))
}

// SBOMGHSAIdempotencyKey uses advisory path in upstream repo.
func SBOMGHSAIdempotencyKey(path string) string {
	return "sbom:ghsa:" + strings.TrimSpace(path)
}

// CoderulesCWEIdempotencyKey is one row from MITRE catalog.
func CoderulesCWEIdempotencyKey(cweID string) string {
	return "coderules:cwe:" + strings.TrimSpace(strings.ToUpper(cweID))
}

// CoderulesSemgrepIdempotencyKey addresses one rule file path in the registry.
func CoderulesSemgrepIdempotencyKey(path string) string {
	return "coderules:semgrep:" + strings.TrimSpace(path)
}

// CoderulesCodeQLIdempotencyKey addresses one .ql path.
func CoderulesCodeQLIdempotencyKey(path string) string {
	return "coderules:codeql:" + strings.TrimSpace(path)
}

// NucleiTemplateIdempotencyKey addresses one template path.
func NucleiTemplateIdempotencyKey(path string) string {
	return "nuclei:template:" + strings.TrimSpace(path)
}

// --- TI idempotency keys (prefix ti:) ---

func TIIoCIdempotencyKey(canonicalIOCID string) string {
	return "ti:ioc:" + strings.TrimSpace(canonicalIOCID)
}

func TIKEVIdempotencyKey(cve string) string {
	return "ti:kev:" + strings.TrimSpace(strings.ToUpper(cve))
}

func TIReportIdempotencyKey(stableReportID string) string {
	return "ti:report:" + strings.TrimSpace(stableReportID)
}

func TICampaignIdempotencyKey(id string) string {
	return "ti:campaign:" + strings.TrimSpace(id)
}

func TIClusterIdempotencyKey(id string) string {
	return "ti:cluster:" + strings.TrimSpace(id)
}

func TIActorIdempotencyKey(actorStableID string) string {
	return "ti:actor:" + strings.TrimSpace(actorStableID)
}

func TILinkCampaignIOCIdempotencyKey(campaignID, iocCanonicalID string) string {
	return "ti:lc:" + strings.TrimSpace(campaignID) + ":" + strings.TrimSpace(iocCanonicalID)
}

func TILinkClusterCampaignIdempotencyKey(clusterID, campaignID string) string {
	return "ti:lkc:" + strings.TrimSpace(clusterID) + ":" + strings.TrimSpace(campaignID)
}

func TILinkCampaignActorIdempotencyKey(campaignID, actorStableID string) string {
	return "ti:lca:" + strings.TrimSpace(campaignID) + ":" + strings.TrimSpace(actorStableID)
}

func TILinkReportMentionsIOCIdempotencyKey(reportID, iocCanonicalID string) string {
	return "ti:lrmi:" + strings.TrimSpace(reportID) + ":" + strings.TrimSpace(iocCanonicalID)
}

func TIJSONLRecordIdempotencyKey(lineSHA256Hex32 string) string {
	return "ti:jsonl:" + strings.TrimSpace(lineSHA256Hex32)
}
