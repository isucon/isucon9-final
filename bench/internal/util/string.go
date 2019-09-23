package util

import (
	crand "crypto/rand"
	"fmt"
)

func SecureRandomStr(b int) (string, error) {
	k := make([]byte, b)
	if _, err := crand.Read(k); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", k), nil
}
