package normalize

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func ReportStableID(link string) string {
	h := sha256.Sum256([]byte("report:" + strings.TrimSpace(link)))
	return hex.EncodeToString(h[:])
}
