// monorepo/native/desktop/maplefile-cli/internal/repo/auth/verifyloginott.go
package authdto

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	dom_authdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/authdto"
)

// loginOTTVerificationDTORepository implements LoginOTTVerificationDTORepository interface
type loginOTTVerificationDTORepository struct {
	logger        *zap.Logger
	configService config.ConfigService
	httpClient    *http.Client
}

// NewLoginOTTVerificationDTORepository creates a new repository for login OTT verification
func NewLoginOTTVerificationDTORepository(logger *zap.Logger, configService config.ConfigService) dom_authdto.LoginOTTVerificationDTORepository {
	logger = logger.Named("LoginOTTVerificationDTORepository")
	return &loginOTTVerificationDTORepository{
		logger:        logger,
		configService: configService,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
	}
}

// VerifyLoginOTT verifies a login one-time token with the server
func (r *loginOTTVerificationDTORepository) VerifyLoginOTT(ctx context.Context, request *dom_authdto.VerifyLoginOTTRequestDTO) (*dom_authdto.VerifyLoginOTTResponseDTO, error) {
	// Get server URL from configuration
	serverURL, err := r.configService.GetCloudProviderAddress(ctx)
	if err != nil {
		return nil, errors.NewAppError("failed to get cloud provider address", err)
	}

	// Convert request to JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, errors.NewAppError("failed to create request", err)
	}

	// Make HTTP request to server
	verifyURL := fmt.Sprintf("%s/iam/api/v1/verify-ott", serverURL)
	// Add emoji to log message
	r.logger.Debug("🌐 Making HTTP request", zap.String("url", verifyURL))

	req, err := http.NewRequestWithContext(ctx, "POST", verifyURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, errors.NewAppError("failed to create HTTP request", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, errors.NewAppError("failed to connect to server", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.NewAppError("failed to read response", err)
	}

	// Check for error status codes
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var errorResponse map[string]interface{}
		if err := json.Unmarshal(body, &errorResponse); err == nil {
			if errMsg, ok := errorResponse["message"].(string); ok {
				return nil, errors.NewAppError(fmt.Sprintf("server error: %s", errMsg), nil)
			}
		}
		return nil, errors.NewAppError(fmt.Sprintf("server returned error status: %s", resp.Status), nil)
	}

	// Parse the response
	var verifyResponse dom_authdto.VerifyLoginOTTResponseDTO
	if err := json.Unmarshal(body, &verifyResponse); err != nil {
		return nil, errors.NewAppError("failed to parse response", err)
	}

	return &verifyResponse, nil
}
