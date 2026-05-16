# Engage assessment reports

Export branded assessment deliverables via `POST /api/visual/export-report`.

## Request

```json
{
  "target": "https://example.com",
  "format": "html",
  "findings": [{"title": "Open port 443", "severity": "medium"}],
  "branding": {
    "organization": "Acme Security",
    "classification": "CONFIDENTIAL",
    "footer": "Authorized assessment only"
  }
}
```

| Field | Values |
|-------|--------|
| `format` | `pdf` (default) or `html` |
| `branding` | Optional org name, classification banner, footer |
| `summary_report` | Alternative to inline `target` + `findings` |

## Response

- **PDF:** `pdf_base64`, `size_bytes`
- **HTML:** `html` body (branded template with findings table)

Set `save_file: true` to persist under `ENGAGE_FILES_DIR`.

## Structured findings (assessment / smart-scan)

Tool stdout is parsed into findings where possible:

| Tool family | Parser | Typical severity |
|-------------|--------|------------------|
| nuclei | JSONL `info.severity` | critical–info |
| nmap | open port lines | info |
| ffuf | `-json` `results[]` or `Status:` lines | 403/401 → medium; 5xx → high |
| sqlmap | injection point blocks | high |

Assessment reports include `executive_summary` with severity counts from aggregated findings.
