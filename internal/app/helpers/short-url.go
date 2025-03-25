package helpers

import (
	"crypto/rand"
	"math/big"
)

const (
	CHARSET = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	LENGTH  = 8
)

func GenerateShortID() string {
	result := make([]byte, LENGTH)
	charsetLength := big.NewInt(int64(len(CHARSET)))

	for i := range result {
		n, err := rand.Int(rand.Reader, charsetLength)
		if err != nil {
			continue
		}
		result[i] = CHARSET[n.Int64()]
	}

	return string(result)
}
