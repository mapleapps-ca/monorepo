// native/desktop/maplefile-cli/internal/domain/recovery/serialization.go
package recovery

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"
)

// Serialize serializes the recovery session into a byte slice using CBOR
func (rs *RecoverySession) Serialize() ([]byte, error) {
	dataBytes, err := cbor.Marshal(rs)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize recovery session: %v", err)
	}
	return dataBytes, nil
}

// NewRecoverySessionFromDeserialized deserializes a recovery session from a byte slice
func NewRecoverySessionFromDeserialized(data []byte) (*RecoverySession, error) {
	// Defensive code: If the input data is empty, return a nil result
	if data == nil {
		return nil, nil
	}

	session := &RecoverySession{}
	if err := cbor.Unmarshal(data, session); err != nil {
		return nil, fmt.Errorf("failed to deserialize recovery session: %v", err)
	}
	return session, nil
}

// Serialize serializes the recovery challenge into a byte slice using CBOR
func (rc *RecoveryChallenge) Serialize() ([]byte, error) {
	dataBytes, err := cbor.Marshal(rc)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize recovery challenge: %v", err)
	}
	return dataBytes, nil
}

// NewRecoveryChallengeFromDeserialized deserializes a recovery challenge from a byte slice
func NewRecoveryChallengeFromDeserialized(data []byte) (*RecoveryChallenge, error) {
	// Defensive code: If the input data is empty, return a nil result
	if data == nil {
		return nil, nil
	}

	challenge := &RecoveryChallenge{}
	if err := cbor.Unmarshal(data, challenge); err != nil {
		return nil, fmt.Errorf("failed to deserialize recovery challenge: %v", err)
	}
	return challenge, nil
}

// Serialize serializes the recovery token into a byte slice using CBOR
func (rt *RecoveryToken) Serialize() ([]byte, error) {
	dataBytes, err := cbor.Marshal(rt)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize recovery token: %v", err)
	}
	return dataBytes, nil
}

// NewRecoveryTokenFromDeserialized deserializes a recovery token from a byte slice
func NewRecoveryTokenFromDeserialized(data []byte) (*RecoveryToken, error) {
	// Defensive code: If the input data is empty, return a nil result
	if data == nil {
		return nil, nil
	}

	token := &RecoveryToken{}
	if err := cbor.Unmarshal(data, token); err != nil {
		return nil, fmt.Errorf("failed to deserialize recovery token: %v", err)
	}
	return token, nil
}

// Serialize serializes the recovery attempt into a byte slice using CBOR
func (ra *RecoveryAttempt) Serialize() ([]byte, error) {
	dataBytes, err := cbor.Marshal(ra)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize recovery attempt: %v", err)
	}
	return dataBytes, nil
}

// NewRecoveryAttemptFromDeserialized deserializes a recovery attempt from a byte slice
func NewRecoveryAttemptFromDeserialized(data []byte) (*RecoveryAttempt, error) {
	// Defensive code: If the input data is empty, return a nil result
	if data == nil {
		return nil, nil
	}

	attempt := &RecoveryAttempt{}
	if err := cbor.Unmarshal(data, attempt); err != nil {
		return nil, fmt.Errorf("failed to deserialize recovery attempt: %v", err)
	}
	return attempt, nil
}
