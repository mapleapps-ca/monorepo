// monorepo/native/desktop/maplefile-cli/internal/domain/localcollection/serialization.go
package collection

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"
)

// Serialize serializes the collection into a byte slice using CBOR
func (c *LocalCollection) Serialize() ([]byte, error) {
	dataBytes, err := cbor.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize local collection: %v", err)
	}
	return dataBytes, nil
}

// NewFromDeserialized deserializes a collection from a byte slice
func NewFromDeserialized(data []byte) (*LocalCollection, error) {
	// Defensive code: If the input data is empty, return a nil result
	if data == nil {
		return nil, nil
	}

	coll := &LocalCollection{}
	if err := cbor.Unmarshal(data, &coll); err != nil {
		return nil, fmt.Errorf("failed to deserialize local collection: %v", err)
	}
	return coll, nil
}
