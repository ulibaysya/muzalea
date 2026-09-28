package rand

import (
	"crypto/rand"
	"encoding/base64"
)

func Base64URL(n int) string {
	bytes := make([]byte, n)
	rand.Read(bytes)
	return base64.RawURLEncoding.EncodeToString(bytes)
}
