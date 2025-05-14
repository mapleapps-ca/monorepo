// monorepo/native/desktop/papercloud-cli/cmd/completelogin/completelogin.go
package completelogin

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/nacl/box"
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

// OTTVerifyResponse stores the response from verify-ott endpoint
type OTTVerifyResponse struct {
	Salt                string `json:"salt"`
	PublicKey           string `json:"publicKey"`
	EncryptedMasterKey  string `json:"encryptedMasterKey"`
	EncryptedPrivateKey string `json:"encryptedPrivateKey"`
	EncryptedChallenge  string `json:"encryptedChallenge"`
	ChallengeID         string `json:"challengeId"`
}

// Constants for cryptographic operations
const (
	NonceSize     = 24
	KeySize       = 32
	PublicKeySize = 32
	SecretKeySize = 32
)

// hexDump prints a hexadecimal dump of a byte slice for debugging
func hexDump(data []byte, label string) {
	fmt.Printf("DEBUG: %s (%d bytes):\n", label, len(data))
	fmt.Printf("  Hex: %s\n", hex.EncodeToString(data))
	// Try to print as string if it looks like ASCII
	isPrintable := true
	for _, b := range data {
		if b < 32 || b > 126 {
			isPrintable = false
			break
		}
	}
	if isPrintable {
		fmt.Printf("  ASCII: %s\n", string(data))
	}
	fmt.Println()
}

func CompleteLoginCmd(configService config.ConfigService, userRepo user.Repository) *cobra.Command {
	var email, password string
	var debugMode bool

	var cmd = &cobra.Command{
		Use:   "completelogin",
		Short: "Finish the login process",
		Long: `
After verifying the one-time token, use this command to complete the login process.

Examples:
  # Complete login with email and password
  papercloud-cli completelogin --email user@example.com --password yourpassword
`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Completing login...")

			if email == "" {
				log.Fatal("Email is required")
			}

			if password == "" {
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

			if debugMode {
				fmt.Println("DEBUG: User data retrieved successfully")
				fmt.Printf("DEBUG: User email: %s\n", userData.Email)
				fmt.Printf("DEBUG: Salt length: %d bytes\n", len(userData.PasswordSalt))
				fmt.Printf("DEBUG: Public key length: %d bytes\n", len(userData.PublicKey.Key))
				fmt.Printf("DEBUG: VerificationID: %s\n", userData.VerificationID)
				fmt.Printf("DEBUG: Encrypted master key nonce length: %d bytes\n", len(userData.EncryptedMasterKey.Nonce))
				fmt.Printf("DEBUG: Encrypted master key ciphertext length: %d bytes\n", len(userData.EncryptedMasterKey.Ciphertext))
				fmt.Printf("DEBUG: Encrypted private key nonce length: %d bytes\n", len(userData.EncryptedPrivateKey.Nonce))
				fmt.Printf("DEBUG: Encrypted private key ciphertext length: %d bytes\n", len(userData.EncryptedPrivateKey.Ciphertext))
			}

			// 2. Check if we have recently completed verifyloginott, which should have saved challenge data
			serverURL, err := configService.GetCloudProviderAddress(ctx)
			if err != nil {
				log.Fatalf("Error loading cloud provider address: %v", err)
				return
			}

			if debugMode {
				fmt.Printf("DEBUG: Server URL: %s\n", serverURL)
			}

			fmt.Printf("DEBUG: Retrieved Encrypted Challenge length: %d bytes\n", len(userData.EncryptedChallenge))

			// Get challenge data directly from the API again
			ottVerifyData := &OTTVerifyResponse{
				Salt:                base64.StdEncoding.EncodeToString(userData.PasswordSalt),
				PublicKey:           base64.StdEncoding.EncodeToString(userData.PublicKey.Key),
				EncryptedMasterKey:  base64.StdEncoding.EncodeToString(append(userData.EncryptedMasterKey.Nonce, userData.EncryptedMasterKey.Ciphertext...)),
				EncryptedPrivateKey: base64.StdEncoding.EncodeToString(append(userData.EncryptedPrivateKey.Nonce, userData.EncryptedPrivateKey.Ciphertext...)),
				EncryptedChallenge:  base64.StdEncoding.EncodeToString(userData.EncryptedChallenge),
				ChallengeID:         userData.VerificationID,
			}

			// Get the challenge ID
			challengeID := ottVerifyData.ChallengeID
			if challengeID == "" {
				log.Fatal("No challenge ID found. Please run verifyloginott first")
			}

			fmt.Printf("Using challenge ID: %s\n", challengeID)
			if debugMode {
				fmt.Printf("DEBUG: Salt (b64): %s\n", ottVerifyData.Salt)
				fmt.Printf("DEBUG: Public Key (b64): %s\n", ottVerifyData.PublicKey)
				fmt.Printf("DEBUG: EncryptedMasterKey (b64): %s\n", ottVerifyData.EncryptedMasterKey)
				fmt.Printf("DEBUG: EncryptedPrivateKey (b64): %s\n", ottVerifyData.EncryptedPrivateKey)
				fmt.Printf("DEBUG: EncryptedChallenge (b64): %s\n", ottVerifyData.EncryptedChallenge)
			}

			// 3. Derive key from password and salt
			salt, err := base64.StdEncoding.DecodeString(ottVerifyData.Salt)
			if err != nil {
				log.Fatalf("Failed to decode salt: %v", err)
			}

			if debugMode {
				hexDump(salt, "Decoded Salt")
			}

			keyEncryptionKey, err := deriveKeyFromPassword(password, salt)
			if err != nil {
				log.Fatalf("Failed to derive key from password: %v", err)
			}
			fmt.Println("Key derived from password successfully")

			if debugMode {
				hexDump(keyEncryptionKey, "Derived Key Encryption Key")
			}

			// 4. Decrypt Master Key using Key Encryption Key
			encryptedMasterKeyCombined, err := base64.StdEncoding.DecodeString(ottVerifyData.EncryptedMasterKey)
			if err != nil {
				log.Fatalf("Failed to decode encrypted master key: %v", err)
			}

			if debugMode {
				hexDump(encryptedMasterKeyCombined, "Encrypted Master Key (Combined)")
			}

			// Split nonce and ciphertext
			if len(encryptedMasterKeyCombined) < NonceSize {
				log.Fatalf("Encrypted master key data is invalid (length: %d, expected minimum: %d)",
					len(encryptedMasterKeyCombined), NonceSize)
			}

			mkNonce := encryptedMasterKeyCombined[:NonceSize]
			mkCiphertext := encryptedMasterKeyCombined[NonceSize:]

			if debugMode {
				hexDump(mkNonce, "Master Key Nonce")
				hexDump(mkCiphertext, "Master Key Ciphertext")
			}

			// Let's trace what's happening in decryptWithSecretBox to debug the "invalid key, nonce, or corrupted ciphertext" error
			if debugMode {
				fmt.Println("DEBUG: Attempting to decrypt master key with secretbox...")
				fmt.Printf("DEBUG: mkNonce length: %d, expected: %d\n", len(mkNonce), NonceSize)
				fmt.Printf("DEBUG: mkCiphertext length: %d\n", len(mkCiphertext))
				fmt.Printf("DEBUG: keyEncryptionKey length: %d, expected: %d\n", len(keyEncryptionKey), KeySize)
			}

			// Try an alternative approach just to see if it works
			if debugMode {
				// This is just for debugging, trying a simpler approach
				fmt.Println("DEBUG: Trying alternative decryption approach...")
				var keyArray [KeySize]byte
				var nonceArray [NonceSize]byte
				copy(keyArray[:], keyEncryptionKey)
				copy(nonceArray[:], mkNonce)
				_, ok := secretbox.Open(nil, mkCiphertext, &nonceArray, &keyArray)
				fmt.Printf("DEBUG: Alternative decryption approach result: %v\n", ok)
			}

			masterKey, err := decryptWithSecretBox(mkCiphertext, mkNonce, keyEncryptionKey)
			if err != nil {
				log.Fatalf("Failed to decrypt master key: %v", err)
			}
			fmt.Println("Master key decrypted successfully")

			if debugMode {
				hexDump(masterKey, "Decrypted Master Key")
			}

			// 5. Decrypt Private Key using Master Key
			encryptedPrivateKeyCombined, err := base64.StdEncoding.DecodeString(ottVerifyData.EncryptedPrivateKey)
			if err != nil {
				log.Fatalf("Failed to decode encrypted private key: %v", err)
			}

			if debugMode {
				hexDump(encryptedPrivateKeyCombined, "Encrypted Private Key (Combined)")
			}

			// Split nonce and ciphertext
			if len(encryptedPrivateKeyCombined) < NonceSize {
				log.Fatalf("Encrypted private key data is invalid (length: %d, expected minimum: %d)",
					len(encryptedPrivateKeyCombined), NonceSize)
			}

			pkNonce := encryptedPrivateKeyCombined[:NonceSize]
			pkCiphertext := encryptedPrivateKeyCombined[NonceSize:]

			if debugMode {
				hexDump(pkNonce, "Private Key Nonce")
				hexDump(pkCiphertext, "Private Key Ciphertext")
			}

			privateKey, err := decryptWithSecretBox(pkCiphertext, pkNonce, masterKey)
			if err != nil {
				log.Fatalf("Failed to decrypt private key: %v", err)
			}
			fmt.Println("Private key decrypted successfully")

			if debugMode {
				hexDump(privateKey, "Decrypted Private Key")
			}

			// 6. Get and decode the public key
			publicKeyBytes, err := base64.StdEncoding.DecodeString(ottVerifyData.PublicKey)
			if err != nil {
				log.Fatalf("Failed to decode public key: %v", err)
			}

			if debugMode {
				hexDump(publicKeyBytes, "Public Key")
			}

			// 7. Get and decode the encrypted challenge
			encryptedChallengeBytes, err := base64.StdEncoding.DecodeString(ottVerifyData.EncryptedChallenge)
			if err != nil {
				log.Fatalf("Failed to decode encrypted challenge: %v", err)
			}

			if debugMode {
				hexDump(encryptedChallengeBytes, "Encrypted Challenge")
			}

			// 8. Decrypt Challenge using Public and Private Keys
			if len(publicKeyBytes) != PublicKeySize {
				log.Fatalf("Invalid public key size: expected %d, got %d", PublicKeySize, len(publicKeyBytes))
			}

			if len(privateKey) != SecretKeySize {
				log.Fatalf("Invalid private key size: expected %d, got %d", SecretKeySize, len(privateKey))
			}

			var pubKeyArray, privKeyArray [32]byte
			copy(pubKeyArray[:], publicKeyBytes)
			copy(privKeyArray[:], privateKey)

			// Decrypt the sealed box challenge
			decryptedChallenge, ok := box.OpenAnonymous(nil, encryptedChallengeBytes, &pubKeyArray, &privKeyArray)
			if !ok {
				log.Fatal("Failed to decrypt challenge: invalid keys or corrupted challenge")
			}
			fmt.Println("Challenge decrypted successfully")

			if debugMode {
				hexDump(decryptedChallenge, "Decrypted Challenge")
			}

			// Try both standard and URL-safe base64 encoding for the server
			decryptedChallengeBase64Std := base64.StdEncoding.EncodeToString(decryptedChallenge)
			decryptedChallengeBase64URL := base64.RawURLEncoding.EncodeToString(decryptedChallenge)

			if debugMode {
				fmt.Printf("DEBUG: Decrypted Challenge (Base64 Standard): %s\n", decryptedChallengeBase64Std)
				fmt.Printf("DEBUG: Decrypted Challenge (Base64 URL): %s\n", decryptedChallengeBase64URL)
			}

			// Try with both Standard and URL-safe base64 encodings
			isSuccess := false
			for i, b64Challenge := range []string{decryptedChallengeBase64Std, decryptedChallengeBase64URL} {
				encType := "Standard"
				if i == 1 {
					encType = "URL-safe"
				}

				if debugMode {
					fmt.Printf("\nDEBUG: Attempting server verification with %s Base64 encoding\n", encType)
				}

				// 9. Send decrypted challenge to server to complete login
				completeLoginReq := CompleteLoginRequest{
					Email:         email,
					ChallengeID:   challengeID,
					DecryptedData: b64Challenge,
				}

				jsonData, err := json.Marshal(completeLoginReq)
				if err != nil {
					log.Fatalf("Error creating request: %v", err)
				}

				if debugMode {
					fmt.Printf("DEBUG: Request payload: %s\n", string(jsonData))
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

				body, err := io.ReadAll(resp.Body)
				resp.Body.Close()

				if err != nil {
					log.Fatalf("Error reading response: %v", err)
				}

				if debugMode {
					fmt.Printf("DEBUG: Server response status: %d\n", resp.StatusCode)
					fmt.Printf("DEBUG: Server response body: %s\n", string(body))
				}

				if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
					isSuccess = true

					// 10. Parse token response
					var tokenResp TokenResponse
					if err := json.Unmarshal(body, &tokenResp); err != nil {
						log.Fatalf("Error parsing token response: %v", err)
					}

					// 11. Update user with tokens
					if err := userRepo.OpenTransaction(); err != nil {
						log.Fatalf("Failed to open transaction: %v", err)
					}

					// Update user with authentication data
					userData.LastLoginAt = time.Now()
					userData.ModifiedAt = time.Now()
					userData.AccessToken = tokenResp.AccessToken
					userData.AccessTokenExpiryTime = tokenResp.AccessTokenExpiryTime
					userData.RefreshToken = tokenResp.RefreshToken
					userData.RefreshTokenExpiryTime = tokenResp.RefreshTokenExpiryTime

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

					// For debugging purposes only.
					fmt.Println("\n✅ Login successful!")
					fmt.Printf("Access Token: %s\n", tokenResp.AccessToken)
					fmt.Printf("Access Token Expires: %s\n", tokenResp.AccessTokenExpiryTime.Format(time.RFC3339))
					fmt.Printf("Refresh Token Expires: %s\n", tokenResp.RefreshTokenExpiryTime.Format(time.RFC3339))

					// Save our authenticated email.
					configService.SetEmail(ctx, email)
					fmt.Println("\n✅ Saved email to our local configuration!")

					break
				} else if i == 0 {
					// If first attempt failed, try URL-safe encoding
					fmt.Println("Standard Base64 encoding failed, trying URL-safe encoding...")
					continue
				} else {
					// Both attempts failed
					log.Fatalf("Server returned error status: %s\nResponse body: %s", resp.Status, string(body))
				}
			}

			if !isSuccess {
				log.Fatal("Login failed with both Base64 encoding formats")
			}
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&email, "email", "e", "", "Email address for the user (required)")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Password for the user (required)")
	cmd.Flags().BoolVarP(&debugMode, "debug", "d", false, "Enable debug output")

	// Mark required flags
	cmd.MarkFlagRequired("email")

	return cmd
}

// getOTTVerificationData attempts to get fresh OTT verification data from the server
func getOTTVerificationData(ctx context.Context, serverURL, email, ott string) (*OTTVerifyResponse, error) {
	// In a real implementation, this would come from a secure storage
	// For now, we'll try to make a new request to verify the OTT
	verifyBody := map[string]string{
		"email": email,
		"ott":   ott,
	}

	jsonData, err := json.Marshal(verifyBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create verify request: %w", err)
	}

	verifyURL := fmt.Sprintf("%s/iam/api/v1/verify-ott", serverURL)
	req, err := http.NewRequest("POST", verifyURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("server returned error: %s", string(body))
	}

	var result OTTVerifyResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse verify response: %w", err)
	}

	return &result, nil
}

// deriveKeyFromPassword derives a key from a password and salt
// This is a simplified version - a real implementation would use Argon2id
func deriveKeyFromPassword(password string, salt []byte) ([]byte, error) {
	if len(salt) != 16 {
		return nil, fmt.Errorf("salt must be 16 bytes, got %d", len(salt))
	}

	// Use the same Argon2id implementation used in registration
	key := argon2.IDKey(
		[]byte(password),
		salt,
		1,           // Argon2OpsLimit (1 iteration as defined in register.go)
		4*1024*1024, // Argon2MemLimit (4 MB as defined in register.go)
		1,           // Argon2Parallelism
		32,          // Argon2KeySize
	)

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
