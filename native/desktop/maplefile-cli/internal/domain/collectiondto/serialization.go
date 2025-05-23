// internal/domain/collectiondto/serialization.go
package collectiondto

import (
	"encoding/json"
	"fmt"
)

// Serialize serializes the collection DTO into a byte slice using JSON
func (f *CollectionDTO) Serialize() ([]byte, error) {
	dataBytes, err := json.Marshal(f)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize local file: %v", err)
	}
	return dataBytes, nil
}

// NewFromDeserialized deserializes a collection DTO from a byte slice
func NewFromDeserialized(data []byte) (*CollectionDTO, error) {
	// Defensive code: If the input data is empty, return a nil result
	if len(data) == 0 { // Check length instead of just nil for empty slices
		return nil, nil
	}

	file := &CollectionDTO{}
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("failed to deserialize collection DTO: %v", err)
	}
	return file, nil
}
