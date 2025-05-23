// internal/domain/remotefile/serialization.go
package remotefile

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"
)

// Serialize serializes the file into a byte slice using CBOR
func (f *RemoteFile) Serialize() ([]byte, error) {
	dataBytes, err := cbor.Marshal(f)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize remote file: %v", err)
	}
	return dataBytes, nil
}

// NewFromDeserialized deserializes a file from a byte slice
func NewFromDeserialized(data []byte) (*RemoteFile, error) {
	// Defensive code: If the input data is empty, return a nil result
	if data == nil {
		return nil, nil
	}

	file := &RemoteFile{}
	if err := cbor.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("failed to deserialize remote file: %v", err)
	}
	return file, nil
}
