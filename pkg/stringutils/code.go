package stringutils

import (
	"bytes"
	"crypto/rand"
	"math/big"
)

const (
	charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// GenerateCode returns a cryptographically random alphanumeric string of the given length.
// Uses crypto/rand (not math/rand) so the output is safe for use as a public short-code.
func GenerateCode(length int) (string, error) {
	var strBuilder bytes.Buffer

	for i := 0; i < length; i++ {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}

		strBuilder.WriteByte(charset[randomIndex.Int64()])
	}

	return strBuilder.String(), nil
}
