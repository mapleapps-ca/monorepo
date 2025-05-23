// internal/domain/file/serialization.go
package file

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"
)

// Serialize serializes the file into a byte slice using CBOR
func (f *File) Serialize() ([]byte, error) {
	dataBytes, err := cbor.Marshal(f)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize local file: %v", err)
	}
	return dataBytes, nil
}

// NewFromDeserialized deserializes a file from a byte slice
func NewFromDeserialized(data []byte) (*File, error) {
	// Defensive code: If the input data is empty, return a nil result
	if data == nil {
		return nil, nil
	}

	file := &File{}
	if err := cbor.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("failed to deserialize local file: %v", err)
	}
	return file, nil
}
