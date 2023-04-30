// Package hash hashes data
package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
)

// Hash make a sha256 hash from data and key
func Hash(src, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(src))
	dst := h.Sum(nil)
	return fmt.Sprintf("%x", dst)
}
