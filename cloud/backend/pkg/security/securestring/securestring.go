package securestring

import (
	"errors"
	"fmt"

	"github.com/awnumar/memguard"
)

// SecureString is used to store a string securely in memory.
type SecureString struct {
	buffer *memguard.LockedBuffer
}

// NewSecureString creates a new SecureString instance from the given string.
func NewSecureString(s string) (*SecureString, error) {
	if len(s) == 0 {
		return nil, errors.New("string cannot be empty")
	}

	// Use memguard's built-in method for creating from bytes
	buffer := memguard.NewBufferFromBytes([]byte(s))

	// Check if buffer was created successfully
	if buffer == nil {
		return nil, errors.New("failed to create buffer")
	}

	return &SecureString{buffer: buffer}, nil
}

// String returns the securely stored string.
func (ss *SecureString) String() string {
	if ss.buffer == nil {
		fmt.Println("String(): buffer is nil")
		return ""
	}
	if !ss.buffer.IsAlive() {
		fmt.Println("String(): buffer is not alive")
		return ""
	}
	return ss.buffer.String()
}

func (ss *SecureString) Bytes() []byte {
	if ss.buffer == nil {
		fmt.Println("Bytes(): buffer is nil")
		return nil
	}
	if !ss.buffer.IsAlive() {
		fmt.Println("Bytes(): buffer is not alive")
		return nil
	}
	return ss.buffer.Bytes()
}

// Wipe removes the string from memory and makes it unrecoverable.
func (ss *SecureString) Wipe() error {

	if ss.buffer != nil {
		if ss.buffer.IsAlive() {
			ss.buffer.Destroy()
		}
	} else {
		// fmt.Println("Wipe(): Buffer is nil")
	}
	ss.buffer = nil
	return nil
}
