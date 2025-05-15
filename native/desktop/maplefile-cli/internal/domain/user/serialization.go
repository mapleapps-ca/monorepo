// monorepo/native/desktop/maplefile-cli/internal/domain/user/serialization.go
package user

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"
)

// Serialize serializes the user into a byte slice.
// This method uses the cbor library to marshal the user into a byte slice.
func (b *User) Serialize() ([]byte, error) {
	// Marshal the user into a byte slice using the cbor library.
	dataBytes, err := cbor.Marshal(b)
	if err != nil {
		// Return an error if the marshaling fails.
		return nil, fmt.Errorf("failed to serialize blockchain sync status: %v", err)
	}
	return dataBytes, nil
}

// NewUserFromDeserialize deserializes an user from a byte slice.
// This method uses the cbor library to unmarshal the byte slice into an user.
func NewUserFromDeserialize(data []byte) (*User, error) {
	// Create a new user variable to return.
	user := &User{}

	// Defensive code: If the input data is empty, return a nil deserialization result.
	if data == nil {
		return nil, nil
	}

	// Unmarshal the byte slice into the user variable using the cbor library.
	if err := cbor.Unmarshal(data, &user); err != nil {
		// Return an error if the unmarshaling fails.
		return nil, fmt.Errorf("failed to deserialize blockchain sync status: %v", err)
	}
	return user, nil
}
