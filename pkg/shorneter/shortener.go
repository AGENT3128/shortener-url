package shorneter

import (
	"crypto/rand"
	"math/big"
)

// Constants for the shortener.
const (
	// CHARSET is the charset for the short ID.
	CHARSET = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	// LENGTH is the length of the short ID.
	LENGTH = 8
)

// GenerateShortID generates a short ID.
func GenerateShortID() (string, error) {
	result := make([]byte, LENGTH)
	charsetLength := big.NewInt(int64(len(CHARSET)))

	for i := range result {
		n, err := rand.Int(rand.Reader, charsetLength)
		if err != nil {
			return "", err
		}
		result[i] = CHARSET[n.Int64()]
	}

	return string(result), nil
}

// GenerateShortIDOptimized generates a short ID optimized for performance.
func GenerateShortIDOptimized() (string, error) {
	result := make([]byte, LENGTH)
	randomBytes := make([]byte, LENGTH)

	// Read all random bytes
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	charsetLen := byte(len(CHARSET))
	for i := range LENGTH {
		// Ensure uniform distribution by rejecting bytes that would create modulo bias
		// Maximum value that can be used without bias
		const maxByteValue = 256
		maxAcceptable := maxByteValue - (maxByteValue % uint16(charsetLen))
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
