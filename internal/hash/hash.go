package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
)

func Hash(src, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(src))
	dst := h.Sum(nil)
	return fmt.Sprintf("%x", dst)
}
