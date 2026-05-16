package audit

import "time"

// Reader supports querying persisted audit events.
type Reader interface {
	Recent(limit int) ([]Event, error)
	ExportNDJSON(since time.Time) ([]byte, error)
}
