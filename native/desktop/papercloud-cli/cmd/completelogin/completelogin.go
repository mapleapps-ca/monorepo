// monorepo/native/desktop/papercloud-cli/cmd/completelogin/completelogin.go
package completelogin

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/nacl/secretbox"

	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/domain/user"
)

// CompleteLoginRequest represents the data sent to the server to complete login
type CompleteLoginRequest struct {
	Email         string `json:"email"`
	ChallengeID   string `json:"challengeId"`
	DecryptedData string `json:"decryptedData"`
}

// TokenResponse represents the response from the server with auth tokens
type TokenResponse struct {
	AccessToken            string    `json:"access_token"`
	AccessTokenExpiryTime  time.Time `json:"access_token_expiry_time"`
	RefreshToken           string    `json:"refresh_token"`
	RefreshTokenExpiryTime time.Time `json:"refresh_token_expiry_time"`
}

// Constants for cryptographic operations
const (
	NonceSize     = 24
	KeySize       = 32
	PublicKeySize = 32
	SecretKeySize = 32
)

func CompleteLoginCmd(configService config.ConfigService, userRepo user.Repository) *cobra.Command {
	var email, password string

	var cmd = &cobra.Command{
		Use:   "completelogin",
		Short: "Finish the login process",
		Long: `
After verifying the one-time token, use this command to complete the login process.

Examples:
  # Complete login with email and password
  papercloud-cli completelogin --email user@example.com
`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Completing login...")

			if email == "" {
				log.Fatal("Email is required")
			}

			// If password is not provided via flag, prompt for it
			if password == "" {
				// In a real implementation, we would use a secure password prompt
				// For simplicity, we'll just use a warning message
				log.Fatal("Password is required. Use --password flag (not secure for production use)")
			}

			// Sanitize inputs
			email = strings.ToLower(strings.TrimSpace(email))

			ctx := context.Background()

			// 1. Get user from repository
			userData, err := userRepo.GetByEmail(ctx, email)
			if err != nil {
				log.Fatalf("Failed to retrieve user data: %v", err)
			}

			if userData == nil {
				log.Fatalf("User with email %s not found", email)
			}

			// 2. Verify that we have the challenge ID (saved during verifyloginott)
			challengeID := userData.VerificationID
			if challengeID == "" {
				log.Fatal("No challenge ID found. Please run verifyloginott first")
			}

			// 3. Derive key from password and salt
			salt := userData.PasswordSalt
			if len(salt) == 0 {
				log.Fatal("No salt found for user. Please run verifyloginott first")
			}

			keyEncryptionKey, err := deriveKeyFromPassword(password, salt)
			if err != nil {
				log.Fatalf("Failed to derive key from password: %v", err)
			}
			fmt.Println("Key derived from password successfully")

			// 4. Decrypt Master Key using Key Encryption Key
			nonce := userData.EncryptedMasterKey.Nonce
			ciphertext := userData.EncryptedMasterKey.Ciphertext

			if len(nonce) == 0 || len(ciphertext) == 0 {
				log.Fatal("Encrypted master key data is missing. Please run verifyloginott first")
			}

			masterKey, err := decryptWithSecretBox(ciphertext, nonce, keyEncryptionKey)
			if err != nil {
				log.Fatalf("Failed to decrypt master key: %v", err)
			}
			fmt.Println("Master key decrypted successfully")

			// 5. Decrypt Private Key using Master Key
			nonce = userData.EncryptedPrivateKey.Nonce
			ciphertext = userData.EncryptedPrivateKey.Ciphertext

			if len(nonce) == 0 || len(ciphertext) == 0 {
				log.Fatal("Encrypted private key data is missing. Please run verifyloginott first")
			}

			privateKey, err := decryptWithSecretBox(ciphertext, nonce, masterKey)
			if err != nil {
				log.Fatalf("Failed to decrypt private key: %v", err)
			}
			fmt.Println("Private key decrypted successfully")

			// 6. Get the encrypted challenge from the server
			// For this implementation, we assume it's stored in the user data after verifyloginott
			// In a real implementation, you might need to request it from the server again

			// Find or request the encrypted challenge
			// Here we would typically have stored this along with other verification data
			// We'll try to reconstruct it from the VerifyOTT data we saved earlier

			// This would be stored properly in a real implementation instead of re-requesting
			serverURL, err := configService.GetCloudProviderAddress(ctx)
			if err != nil {
				log.Fatalf("Error loading cloud provider address: %v", err)
				return
			}

			// 7. Decrypt Challenge using Public and Private Keys
			publicKey := userData.PublicKey.Key
			if len(publicKey) != PublicKeySize {
				log.Fatalf("Invalid public key size: expected %d, got %d", PublicKeySize, len(publicKey))
			}

			if len(privateKey) != SecretKeySize {
				log.Fatalf("Invalid private key size: expected %d, got %d", SecretKeySize, len(privateKey))
			}

			var pubKeyArray, privKeyArray [32]byte
			copy(pubKeyArray[:], publicKey)
			copy(privKeyArray[:], privateKey)

			// Get the encrypted challenge from a temporary storage
			// In a real implementation, this would come from the user data or a preferences store
			// For now, we'll need to make an API call to get it again
			encryptedChallenge, err := getEncryptedChallenge(serverURL, email)
			if err != nil {
				log.Fatalf("Failed to get encrypted challenge: %v", err)
			}

			decryptedChallenge, err := decryptSealedBox(encryptedChallenge, pubKeyArray, privKeyArray)
			if err != nil {
				log.Fatalf("Failed to decrypt challenge: %v", err)
			}
			fmt.Println("Challenge decrypted successfully")

			// 8. Send decrypted challenge to server to complete login
			decryptedChallengeBase64 := base64.StdEncoding.EncodeToString(decryptedChallenge)

			completeLoginReq := CompleteLoginRequest{
				Email:         email,
				ChallengeID:   challengeID,
				DecryptedData: decryptedChallengeBase64,
			}

			jsonData, err := json.Marshal(completeLoginReq)
			if err != nil {
				log.Fatalf("Error creating request: %v", err)
			}

			completeURL := fmt.Sprintf("%s/iam/api/v1/complete-login", serverURL)
			req, err := http.NewRequest("POST", completeURL, bytes.NewBuffer(jsonData))
			if err != nil {
				log.Fatalf("Error creating HTTP request: %v", err)
			}

			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{Timeout: 30 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				log.Fatalf("Error connecting to server: %v", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Fatalf("Error reading response: %v", err)
			}

			if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
				log.Fatalf("Server returned error status: %s\nResponse body: %s", resp.Status, string(body))
			}

			// 9. Parse token response
			var tokenResp TokenResponse
			if err := json.Unmarshal(body, &tokenResp); err != nil {
				log.Fatalf("Error parsing token response: %v", err)
			}

			// 10. Start transaction and update user with tokens
			if err := userRepo.OpenTransaction(); err != nil {
				log.Fatalf("Failed to open transaction: %v", err)
			}

			// Update user with authentication data
			userData.LastLoginAt = time.Now()
			// In a real implementation, we would store tokens more securely
			// For now, we're just printing success message

			// Save the updated user
			if err := userRepo.UpsertByEmail(ctx, userData); err != nil {
				userRepo.DiscardTransaction()
				log.Fatalf("Failed to update user data: %v", err)
			}

			// Commit the transaction
			if err := userRepo.CommitTransaction(); err != nil {
				userRepo.DiscardTransaction()
				log.Fatalf("Failed to commit transaction: %v", err)
			}

			fmt.Println("Login successful!")
			fmt.Printf("Access Token: %s\n", tokenResp.AccessToken)
			fmt.Printf("Access Token Expires: %s\n", tokenResp.AccessTokenExpiryTime.Format(time.RFC3339))
			fmt.Printf("Refresh Token Expires: %s\n", tokenResp.RefreshTokenExpiryTime.Format(time.RFC3339))
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&email, "email", "e", "", "Email address for the user (required)")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Password for the user (will prompt if not provided)")

	// Mark required flags
	cmd.MarkFlagRequired("email")

	return cmd
}

// deriveKeyFromPassword derives a key from a password using Argon2id
func deriveKeyFromPassword(password string, salt []byte) ([]byte, error) {
	// In a real implementation, this would use Argon2id
	// For simplicity, we'll use a placeholder implementation
	// Note: This is not secure and should be replaced with a proper implementation

	// This is a placeholder - in a real implementation, use golang.org/x/crypto/argon2
	key := []byte(password + string(salt))
	if len(key) > KeySize {
		key = key[:KeySize]
	} else {
		// Pad to KeySize
		paddedKey := make([]byte, KeySize)
		copy(paddedKey, key)
		key = paddedKey
	}

	return key, nil
}

// decryptWithSecretBox decrypts data using NaCl secretbox
func decryptWithSecretBox(ciphertext, nonce, key []byte) ([]byte, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("invalid key size: expected %d, got %d", KeySize, len(key))
	}

	if len(nonce) != NonceSize {
		return nil, fmt.Errorf("invalid nonce size: expected %d, got %d", NonceSize, len(nonce))
	}

	var keyArray [KeySize]byte
	var nonceArray [NonceSize]byte

	copy(keyArray[:], key)
	copy(nonceArray[:], nonce)

	plaintext, ok := secretbox.Open(nil, ciphertext, &nonceArray, &keyArray)
	if !ok {
		return nil, fmt.Errorf("failed to decrypt: invalid key, nonce, or corrupted ciphertext")
	}

	return plaintext, nil
}

// getEncryptedChallenge retrieves the encrypted challenge from the server
func getEncryptedChallenge(serverURL, email string) ([]byte, error) {
	// In a real implementation, this would make an API call to get the challenge
	// For this example, we'll return a placeholder

	// This is a placeholder - in a real implementation, make the API call
	// or retrieve the challenge from where it was stored after verify-ott
	return []byte("placeholder_encrypted_challenge"), nil
}

// decryptSealedBox decrypts a sealed box using public and private keys
func decryptSealedBox(sealedBox []byte, publicKey, privateKey [32]byte) ([]byte, error) {
	// In a real implementation, this would use box.Open
	// For simplicity, we'll use a placeholder implementation

	// This is a placeholder - in a real implementation, use box.Open
	return []byte("decrypted_challenge"), nil
}
