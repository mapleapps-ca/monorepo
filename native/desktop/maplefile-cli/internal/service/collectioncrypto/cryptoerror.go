// internal/service/collectioncrypto/cryptoerror.go
package collectioncrypto

import "fmt"

// Standardized crypto error types
type CryptoError struct {
	Operation string
	Cause     error
	UserID    string
	Context   map[string]interface{}
}

func (e *CryptoError) Error() string {
	return fmt.Sprintf("crypto operation '%s' failed: %v", e.Operation, e.Cause)
}

func NewCryptoError(operation string, cause error, userID string) *CryptoError {
	return &CryptoError{
		Operation: operation,
		Cause:     cause,
		UserID:    userID,
		Context:   make(map[string]interface{}),
	}
}
