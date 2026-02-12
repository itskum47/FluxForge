package main

import (
	"crypto/rand"
	"fmt"
	"io"
)

// generateUUID generates a random UUID-like string.
func generateUUID() string {
	b := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	b[8] = b[8]&0x3f | 0x80
	b[6] = b[6]&0x0f | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
