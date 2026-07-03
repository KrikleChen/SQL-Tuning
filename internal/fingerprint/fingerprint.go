package fingerprint

import (
	"crypto/sha256"
	"encoding/hex"
)

func SHA256(input string) string {
	sum := sha256.Sum256([]byte(input))
	return "sha256:" + hex.EncodeToString(sum[:])
}
