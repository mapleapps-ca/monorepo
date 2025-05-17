// internal/domain/collection/serialization.go
package collection

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"
)

// Serialize serializes the collection into a byte slice using CBOR
func (c *Collection) Serialize() ([]byte, error) {
	dataBytes, err := cbor.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize collection: %v", err)
	}
	return dataBytes, nil
}

// NewFromDeserialized deserializes a collection from a byte slice
func NewFromDeserialized(data []byte) (*Collection, error) {
	// Defensive code: If the input data is empty, return a nil result
	if data == nil {
		return nil, nil
	}

	coll := &Collection{}
	if err := cbor.Unmarshal(data, &coll); err != nil {
		return nil, fmt.Errorf("failed to deserialize collection: %v", err)
	}
	return coll, nil
}
