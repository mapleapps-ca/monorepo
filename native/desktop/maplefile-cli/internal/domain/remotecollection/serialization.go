// monorepo/native/desktop/maplefile-cli/internal/domain/remotecollection/serialization.go
package remotecollection

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"
)

// Serialize serializes the collection into a byte slice using CBOR
func (c *RemoteCollection) Serialize() ([]byte, error) {
	dataBytes, err := cbor.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("failed to remote collection: %v", err)
	}
	return dataBytes, nil
}

// NewFromDeserialized deserializes a collection from a byte slice
func NewFromDeserialized(data []byte) (*RemoteCollection, error) {
	// Defensive code: If the input data is empty, return a nil result
	if data == nil {
		return nil, nil
	}

	coll := &RemoteCollection{}
	if err := cbor.Unmarshal(data, &coll); err != nil {
		return nil, fmt.Errorf("failed to deserialize remote collection: %v", err)
	}
	return coll, nil
}
