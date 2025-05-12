package random

import (
	"crypto/rand"
	"math/big"
)

// GenerateSixDigitCode generates a cryptographically secure random 6-digit number
func GenerateSixDigitCode() (string, error) {
	// Generate a random number between 100000 and 999999
	max := big.NewInt(900000) // 999999 - 100000 + 1
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	// Add 100000 to ensure 6 digits
	n.Add(n, big.NewInt(100000))

	return n.String(), nil
}
