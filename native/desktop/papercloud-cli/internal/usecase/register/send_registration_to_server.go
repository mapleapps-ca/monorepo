// internal/usecase/register/send_registration_to_server.go
package register

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/domain/user"
)

// RegisterRequest represents the data structure needed for user registration with the server
type RegisterRequest struct {
	// Personal information
	BetaAccessCode                                 string `json:"beta_access_code"`
	FirstName                                      string `json:"first_name"`
	LastName                                       string `json:"last_name"`
	Email                                          string `json:"email"`
	Phone                                          string `json:"phone,omitempty"`
	Country                                        string `json:"country,omitempty"`
	CountryOther                                   string `json:"country_other,omitempty"`
	Timezone                                       string `json:"timezone"`
	AgreeTermsOfService                            bool   `json:"agree_terms_of_service"`
	AgreePromotions                                bool   `json:"agree_promotions,omitempty"`
	AgreeToTrackingAcrossThirdPartyAppsAndServices bool   `json:"agree_to_tracking_across_third_party_apps_and_services,omitempty"`
	Module                                         int    `json:"module"`

	// E2EE related fields
	Salt                              string `json:"salt"`
	PublicKey                         string `json:"publicKey"`
	EncryptedMasterKey                string `json:"encryptedMasterKey"`
	EncryptedPrivateKey               string `json:"encryptedPrivateKey"`
	EncryptedRecoveryKey              string `json:"encryptedRecoveryKey"`
	MasterKeyEncryptedWithRecoveryKey string `json:"masterKeyEncryptedWithRecoveryKey"`
	VerificationID                    string `json:"verificationID"`
}

// SendRegistrationToServerInput contains the input for sending registration to server
type SendRegistrationToServerInput struct {
	User           *user.User
	BetaAccessCode string
	Module         int
}

// SendRegistrationToServerUseCase defines the interface for sending registration to server
type SendRegistrationToServerUseCase interface {
	Execute(ctx context.Context, input SendRegistrationToServerInput) (string, error)
}

type sendRegistrationToServerUseCase struct {
	configService config.ConfigService
}

// NewSendRegistrationToServerUseCase creates a new SendRegistrationToServerUseCase
func NewSendRegistrationToServerUseCase(configService config.ConfigService) SendRegistrationToServerUseCase {
	return &sendRegistrationToServerUseCase{
		configService: configService,
	}
}

// Execute sends registration data to the server
func (uc *sendRegistrationToServerUseCase) Execute(ctx context.Context, input SendRegistrationToServerInput) (string, error) {
	// Get the server URL from configuration
	serverURL, err := uc.configService.GetCloudProviderAddress(ctx)
	if err != nil {
		return "", fmt.Errorf("error loading cloud provider address: %w", err)
	}

	// Prepare data for the server
	saltBase64 := base64.RawURLEncoding.EncodeToString(input.User.PasswordSalt)
	publicKeyBase64 := base64.RawURLEncoding.EncodeToString(input.User.PublicKey.Key)

	// Combine nonce and ciphertext for encrypted data
	encryptedMasterKeyBytes := append(input.User.EncryptedMasterKey.Nonce, input.User.EncryptedMasterKey.Ciphertext...)
	encryptedPrivateKeyBytes := append(input.User.EncryptedPrivateKey.Nonce, input.User.EncryptedPrivateKey.Ciphertext...)
	encryptedRecoveryKeyBytes := append(input.User.EncryptedRecoveryKey.Nonce, input.User.EncryptedRecoveryKey.Ciphertext...)
	masterKeyEncryptedWithRecoveryKeyBytes := append(
		input.User.MasterKeyEncryptedWithRecoveryKey.Nonce,
		input.User.MasterKeyEncryptedWithRecoveryKey.Ciphertext...,
	)

	// Convert to base64 for API
	encryptedMasterKeyBase64 := base64.RawURLEncoding.EncodeToString(encryptedMasterKeyBytes)
	encryptedPrivateKeyBase64 := base64.RawURLEncoding.EncodeToString(encryptedPrivateKeyBytes)
	encryptedRecoveryKeyBase64 := base64.RawURLEncoding.EncodeToString(encryptedRecoveryKeyBytes)
	masterKeyEncryptedWithRecoveryKeyBase64 := base64.RawURLEncoding.EncodeToString(masterKeyEncryptedWithRecoveryKeyBytes)

	// Create registration request
	registerReq := RegisterRequest{
		Email:               strings.ToLower(input.User.Email),
		FirstName:           input.User.FirstName,
		LastName:            input.User.LastName,
		Phone:               input.User.Phone,
		Country:             input.User.Country,
		Timezone:            input.User.Timezone,
		BetaAccessCode:      input.BetaAccessCode,
		Module:              input.Module,
		AgreeTermsOfService: input.User.AgreeTermsOfService,
		AgreePromotions:     input.User.AgreePromotions,
		AgreeToTrackingAcrossThirdPartyAppsAndServices: input.User.AgreeToTrackingAcrossThirdPartyAppsAndServices,

		// E2EE fields
		Salt:                              saltBase64,
		PublicKey:                         publicKeyBase64,
		EncryptedMasterKey:                encryptedMasterKeyBase64,
		EncryptedPrivateKey:               encryptedPrivateKeyBase64,
		EncryptedRecoveryKey:              encryptedRecoveryKeyBase64,
		MasterKeyEncryptedWithRecoveryKey: masterKeyEncryptedWithRecoveryKeyBase64,
		VerificationID:                    input.User.VerificationID,
	}

	// Convert request to JSON
	jsonData, err := json.Marshal(registerReq)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	// Make HTTP request to server
	registerURL := fmt.Sprintf("%s/iam/api/v1/register", serverURL)

	// Create and execute the HTTP request
	req, err := http.NewRequest("POST", registerURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error connecting to server: %w", err)
	}
	defer resp.Body.Close()

	// Read and process the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %w", err)
	}

	// Check response status code
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server returned error status: %s\nResponse body: %s", resp.Status, string(body))
	}

	return "Registration successful! Please check your email for verification instructions.", nil
}
