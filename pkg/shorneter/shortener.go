package shorneter

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

func GenerateShortIDOptimized() (string, error) {
	result := make([]byte, LENGTH)
	randomBytes := make([]byte, LENGTH)

	// Read all random bytes
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	charsetLen := byte(len(CHARSET))
	for i := 0; i < LENGTH; i++ {
		// Ensure uniform distribution by rejecting bytes that would create modulo bias
		// Maximum value that can be used without bias
		maxAcceptable := 256 - (256 % uint16(charsetLen))
		b := randomBytes[i]

		// Reject values that would create bias
		for b >= byte(maxAcceptable) {
			newByte := make([]byte, 1)
			if _, err := rand.Read(newByte); err != nil {
				return "", err
			}
			b = newByte[0]
		}

		// Map the random byte to our charset
		result[i] = CHARSET[b%charsetLen]
	}

	return string(result), nil
}
