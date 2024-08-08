package utils

import (
	"crypto/sha256"
	"fmt"
)

func Hash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))

	hbytes := h.Sum(nil)

	return fmt.Sprintf("%x", hbytes)
}
