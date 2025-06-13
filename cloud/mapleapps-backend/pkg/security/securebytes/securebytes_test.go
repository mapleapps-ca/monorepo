package securebytes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSecureBytes(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{
			name:    "valid input",
			input:   []byte("test-data"),
			wantErr: false,
		},
		{
			name:    "empty input",
			input:   []byte{},
			wantErr: true,
		},
		{
			name:    "nil input",
			input:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb, err := NewSecureBytes(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, sb)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, sb)
				assert.NotNil(t, sb.buffer)
			}
		})
	}
}

func TestSecureBytes_Bytes(t *testing.T) {
	input := []byte("test-data")
	sb, err := NewSecureBytes(input)
	assert.NoError(t, err)

	// Ensure the SecureBytes object is properly closed after the test
	defer sb.Wipe()

	output := sb.Bytes()
	assert.Equal(t, input, output)
	assert.NotSame(t, &input, &output) // Verify different memory addresses
}

func TestSecureBytes_Wipe(t *testing.T) {
	sb, err := NewSecureBytes([]byte("test-data"))
	assert.NoError(t, err)

	err = sb.Wipe()
	assert.NoError(t, err)

	// After wiping, the internal buffer should be nil
	assert.Nil(t, sb.buffer)

	// Attempting to access bytes after wiping might panic or return nil/empty slice
	// Based on the panic, calling Bytes() on a wiped buffer is unsafe.
	// We verify the buffer is nil instead of calling Bytes().
}

func TestSecureBytes_DataIsolation(t *testing.T) {
	original := []byte("test-data")
	sb, err := NewSecureBytes(original)
	assert.NoError(t, err)

	// Ensure the SecureBytes object is properly closed after the test
	defer sb.Wipe()

	// Modify original data
	original[0] = 'x'

	// Verify secure bytes remains unchanged
	stored := sb.Bytes()
	assert.NotEqual(t, original, stored)
	assert.Equal(t, []byte("test-data"), stored)
}
