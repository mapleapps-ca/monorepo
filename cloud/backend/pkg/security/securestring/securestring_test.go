package securestring

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSecureString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid string",
			input:   "test-string",
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss, err := NewSecureString(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, ss)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, ss)
				assert.NotNil(t, ss.buffer)
			}
		})
	}
}

func TestSecureString_String(t *testing.T) {
	input := "test-string"
	ss, err := NewSecureString(input)
	assert.NoError(t, err)

	output := ss.String()
	assert.Equal(t, input, output)
}

func TestSecureString_Wipe(t *testing.T) {
	ss, err := NewSecureString("test-string")
	assert.NoError(t, err)

	err = ss.Wipe()
	assert.NoError(t, err)
	assert.Nil(t, ss.buffer)

	// Verify string is wiped
	output := ss.String()
	assert.Empty(t, output)
}

func TestSecureString_DataIsolation(t *testing.T) {
	original := "test-string"
	ss, err := NewSecureString(original)
	assert.NoError(t, err)

	// Attempt to modify original
	original = "modified"

	// Verify secure string remains unchanged
	stored := ss.String()
	assert.NotEqual(t, original, stored)
	assert.Equal(t, "test-string", stored)
}

func TestSecureString_StringConsistency(t *testing.T) {
	input := "test-string"
	ss, err := NewSecureString(input)
	assert.NoError(t, err)

	// Multiple calls should return same value
	assert.Equal(t, ss.String(), ss.String())
}
