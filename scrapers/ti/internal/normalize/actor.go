package normalize

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// ActorStableID returns a deterministic id for an actor name (lowercased, trimmed).
func ActorStableID(name string) string {
	n := strings.TrimSpace(strings.ToLower(name))
	h := sha256.Sum256([]byte("actor:" + n))
	return hex.EncodeToString(h[:])
}
