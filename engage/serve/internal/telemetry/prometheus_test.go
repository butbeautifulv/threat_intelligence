package telemetry

import "testing"

func TestRecordToolRun(t *testing.T) {
	RecordToolRun("nmap_scan", true)
	RecordToolRun("nmap_scan", false)
	RecordAuditEvent()
	SetJobsPending(2)
	SetCacheEntries(5)
}
