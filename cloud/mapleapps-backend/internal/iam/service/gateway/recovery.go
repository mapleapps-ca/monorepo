// monorepo/cloud/mapleapps-backend/internal/iam/service/gateway/recovery.go
package gateway

import (
	"context"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_recovery "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/recovery"
	uc_recovery "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/usecase/recovery"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

// InitiateRecoveryRequestDTO represents the request to initiate account recovery
type InitiateRecoveryRequestDTO struct {
	Email  string `json:"email"`
	Method string `json:"method,omitempty"` // Default: "recovery_key"
}

// InitiateRecoveryResponseDTO represents the response from initiating recovery
type InitiateRecoveryResponseDTO struct {
	SessionID          string `json:"session_id"`
	ChallengeID        string `json:"challenge_id"`
	EncryptedChallenge string `json:"encrypted_challenge"`
	ExpiresIn          int    `json:"expires_in"` // seconds
}

// VerifyRecoveryRequestDTO represents the request to verify recovery challenge
type VerifyRecoveryRequestDTO struct {
	SessionID          string `json:"session_id"`
	DecryptedChallenge string `json:"decrypted_challenge"`
}

// VerifyRecoveryResponseDTO represents the response from verifying recovery
type VerifyRecoveryResponseDTO struct {
	RecoveryToken            string `json:"recovery_token"`
	MasterKeyWithRecoveryKey string `json:"master_key_encrypted_with_recovery_key"`
	ExpiresIn                int    `json:"expires_in"` // seconds
}

// CompleteRecoveryRequestDTO represents the request to complete recovery
type CompleteRecoveryRequestDTO struct {
	RecoveryToken               string `json:"recovery_token"`
	NewSalt                     string `json:"new_salt"`
	NewEncryptedMasterKey       string `json:"new_encrypted_master_key"`
	NewEncryptedPrivateKey      string `json:"new_encrypted_private_key"`
	NewEncryptedRecoveryKey     string `json:"new_encrypted_recovery_key"`
	NewMasterKeyWithRecoveryKey string `json:"new_master_key_encrypted_with_recovery_key"`
}

// CompleteRecoveryResponseDTO represents the response from completing recovery
type CompleteRecoveryResponseDTO struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// Recovery Services
type InitiateRecoveryService interface {
	Execute(ctx context.Context, req *InitiateRecoveryRequestDTO) (*InitiateRecoveryResponseDTO, error)
}

type VerifyRecoveryService interface {
	Execute(ctx context.Context, req *VerifyRecoveryRequestDTO) (*VerifyRecoveryResponseDTO, error)
}

type CompleteRecoveryService interface {
	Execute(ctx context.Context, req *CompleteRecoveryRequestDTO) (*CompleteRecoveryResponseDTO, error)
}

// InitiateRecoveryService Implementation
type initiateRecoveryServiceImpl struct {
	config  *config.Configuration
	logger  *zap.Logger
	useCase uc_recovery.InitiateRecoveryUseCase
}

func NewInitiateRecoveryService(
	config *config.Configuration,
	logger *zap.Logger,
	useCase uc_recovery.InitiateRecoveryUseCase,
) InitiateRecoveryService {
	logger = logger.Named("InitiateRecoveryService")
	return &initiateRecoveryServiceImpl{
		config:  config,
		logger:  logger,
		useCase: useCase,
	}
}

func (s *initiateRecoveryServiceImpl) Execute(ctx context.Context, req *InitiateRecoveryRequestDTO) (*InitiateRecoveryResponseDTO, error) {
	// Validate input
	e := make(map[string]string)
	if req.Email == "" {
		e["email"] = "Email address is required"
	}
	if len(e) != 0 {
		return nil, httperror.NewForBadRequest(&e)
	}

	// Sanitize input
	req.Email = strings.ToLower(req.Email)
	req.Email = strings.ReplaceAll(req.Email, " ", "")

	// Default method
	method := dom_recovery.RecoveryMethodRecoveryKey
	if req.Method != "" {
		method = dom_recovery.RecoveryMethod(req.Method)
	}

	// Execute use case
	result, err := s.useCase.Execute(ctx, req.Email, method)
	if err != nil {
		return nil, err
	}

	// Calculate expires in seconds
	expiresIn := int(result.ExpiresAt.Sub(result.CreatedAt).Seconds())

	return &InitiateRecoveryResponseDTO{
		SessionID:          result.SessionID,
		ChallengeID:        result.ChallengeID,
		EncryptedChallenge: result.EncryptedChallenge,
		ExpiresIn:          expiresIn,
	}, nil
}

// VerifyRecoveryService Implementation
type verifyRecoveryServiceImpl struct {
	config  *config.Configuration
	logger  *zap.Logger
	useCase uc_recovery.VerifyRecoveryUseCase
}

func NewVerifyRecoveryService(
	config *config.Configuration,
	logger *zap.Logger,
	useCase uc_recovery.VerifyRecoveryUseCase,
) VerifyRecoveryService {
	logger = logger.Named("VerifyRecoveryService")
	return &verifyRecoveryServiceImpl{
		config:  config,
		logger:  logger,
		useCase: useCase,
	}
}

func (s *verifyRecoveryServiceImpl) Execute(ctx context.Context, req *VerifyRecoveryRequestDTO) (*VerifyRecoveryResponseDTO, error) {
	// Validate input
	e := make(map[string]string)
	if req.SessionID == "" {
		e["session_id"] = "Session ID is required"
	}
	if req.DecryptedChallenge == "" {
		e["decrypted_challenge"] = "Decrypted challenge is required"
	}
	if len(e) != 0 {
		return nil, httperror.NewForBadRequest(&e)
	}

	// Execute use case
	ucReq := &uc_recovery.VerifyRecoveryRequest{
		SessionID:          req.SessionID,
		DecryptedChallenge: req.DecryptedChallenge,
	}

	result, err := s.useCase.Execute(ctx, ucReq)
	if err != nil {
		return nil, err
	}

	// Calculate expires in seconds
	expiresIn := int(result.ExpiresAt.Sub(time.Now()).Seconds())

	return &VerifyRecoveryResponseDTO{
		RecoveryToken:            result.RecoveryToken,
		MasterKeyWithRecoveryKey: result.MasterKeyWithRecoveryKey,
		ExpiresIn:                expiresIn,
	}, nil
}

// CompleteRecoveryService Implementation
type completeRecoveryServiceImpl struct {
	config  *config.Configuration
	logger  *zap.Logger
	useCase uc_recovery.CompleteRecoveryUseCase
}

func NewCompleteRecoveryService(
	config *config.Configuration,
	logger *zap.Logger,
	useCase uc_recovery.CompleteRecoveryUseCase,
) CompleteRecoveryService {
	logger = logger.Named("CompleteRecoveryService")
	return &completeRecoveryServiceImpl{
		config:  config,
		logger:  logger,
		useCase: useCase,
	}
}

func (s *completeRecoveryServiceImpl) Execute(ctx context.Context, req *CompleteRecoveryRequestDTO) (*CompleteRecoveryResponseDTO, error) {
	// Validate input
	e := make(map[string]string)
	if req.RecoveryToken == "" {
		e["recovery_token"] = "Recovery token is required"
	}
	if req.NewSalt == "" {
		e["new_salt"] = "New salt is required"
	}
	if req.NewEncryptedMasterKey == "" {
		e["new_encrypted_master_key"] = "New encrypted master key is required"
	}
	if req.NewEncryptedPrivateKey == "" {
		e["new_encrypted_private_key"] = "New encrypted private key is required"
	}
	if req.NewEncryptedRecoveryKey == "" {
		e["new_encrypted_recovery_key"] = "New encrypted recovery key is required"
	}
	if req.NewMasterKeyWithRecoveryKey == "" {
		e["new_master_key_encrypted_with_recovery_key"] = "New master key encrypted with recovery key is required"
	}
	if len(e) != 0 {
		return nil, httperror.NewForBadRequest(&e)
	}

	// Execute use case
	ucReq := &uc_recovery.CompleteRecoveryRequest{
		RecoveryToken:               req.RecoveryToken,
		NewSalt:                     req.NewSalt,
		NewEncryptedMasterKey:       req.NewEncryptedMasterKey,
		NewEncryptedPrivateKey:      req.NewEncryptedPrivateKey,
		NewEncryptedRecoveryKey:     req.NewEncryptedRecoveryKey,
		NewMasterKeyWithRecoveryKey: req.NewMasterKeyWithRecoveryKey,
	}

	result, err := s.useCase.Execute(ctx, ucReq)
	if err != nil {
		return nil, err
	}

	return &CompleteRecoveryResponseDTO{
		Success: result.Success,
		Message: result.Message,
	}, nil
}
