package idempotency

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// GenerateKey builds a deterministic key using all provided parts.
func GenerateKey(parts ...interface{}) string {
	h := sha256.New()
	for _, part := range parts {
		fmt.Fprintf(h, "%v:", part)
	}

	return hex.EncodeToString(h.Sum(nil))
}
