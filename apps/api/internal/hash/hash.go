package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
)

// ComputeSHA256 computes the SHA-256 hash of the content from the reader.
// The reader is consumed entirely during this operation.
// Returns the hash as a lowercase hexadecimal string (64 characters).
func ComputeSHA256(r io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", fmt.Errorf("failed to compute hash: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// ComputeSHA256FromBytes computes the SHA-256 hash from a byte slice.
// Returns the hash as a lowercase hexadecimal string (64 characters).
func ComputeSHA256FromBytes(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// ValidateHash checks if a hash string is a valid SHA-256 hash.
// A valid hash must be exactly 64 hexadecimal characters.
func ValidateHash(hash string) bool {
	if len(hash) != 64 {
		return false
	}
	for _, c := range hash {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
			return false
		}
	}
	return true
}
