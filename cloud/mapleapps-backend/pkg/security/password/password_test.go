package password

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sstring "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/security/securestring"
)

func TestPasswordHashing(t *testing.T) {
	t.Log("TestPasswordHashing: Starting")

	provider := NewProvider()
	t.Log("TestPasswordHashing: Provider created")

	password, err := sstring.NewSecureString("test-password")
	require.NoError(t, err)
	t.Log("TestPasswordHashing: Password SecureString created")
	fmt.Println("TestPasswordHashing: Password SecureString created")

	// Let's add a timeout to see if we can pinpoint the issue
	done := make(chan bool)
	go func() {
		fmt.Println("TestPasswordHashing: Generating hash...")
		hash, err := provider.GenerateHashFromPassword(password)
		fmt.Printf("TestPasswordHashing: Hash generated: %v, error: %v\n", hash != "", err)

		if err == nil {
			fmt.Println("TestPasswordHashing: Comparing password and hash...")
			match, err := provider.ComparePasswordAndHash(password, hash)
			fmt.Printf("TestPasswordHashing: Comparison done: match=%v, error=%v\n", match, err)
		}

		done <- true
	}()

	select {
	case <-done:
		fmt.Println("TestPasswordHashing: Test completed successfully")
	case <-time.After(10 * time.Second):
		t.Fatal("Test timed out after 10 seconds")
	}

	fmt.Println("TestPasswordHashing: Cleaning up password...")
	password.Wipe()
	fmt.Println("TestPasswordHashing: Done")
}
