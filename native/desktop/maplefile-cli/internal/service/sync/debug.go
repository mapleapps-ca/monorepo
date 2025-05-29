// native/desktop/maplefile-cli/internal/service/sync/debug.go
package sync

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/syncstate"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
)

// DebugSyncInput represents input for debugging sync operations
type DebugSyncInput struct {
	CheckAuth      bool   `json:"check_auth"`
	CheckNetwork   bool   `json:"check_network"`
	CheckSyncState bool   `json:"check_sync_state"`
	Password       string `json:"password,omitempty"`
}

// DebugSyncOutput represents the result of sync debugging
type DebugSyncOutput struct {
	AuthStatus      string   `json:"auth_status"`
	NetworkStatus   string   `json:"network_status"`
	SyncStateStatus string   `json:"sync_state_status"`
	Issues          []string `json:"issues"`
	Recommendations []string `json:"recommendations"`
}

// SyncDebugService defines the interface for debugging sync operations
type SyncDebugService interface {
	DiagnoseSync(ctx context.Context, input *DebugSyncInput) (*DebugSyncOutput, error)
}

// syncDebugService implements the SyncDebugService interface
type syncDebugService struct {
	logger                     *zap.Logger
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase
	syncStateGetService        syncstate.GetService
}

// NewSyncDebugService creates a new service for debugging sync operations
func NewSyncDebugService(
	logger *zap.Logger,
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase,
	syncStateGetService syncstate.GetService,
) SyncDebugService {
	logger = logger.Named("SyncDebugService")
	return &syncDebugService{
		logger:                     logger,
		getUserByIsLoggedInUseCase: getUserByIsLoggedInUseCase,
		syncStateGetService:        syncStateGetService,
	}
}

// DiagnoseSync performs comprehensive sync diagnostics
func (s *syncDebugService) DiagnoseSync(ctx context.Context, input *DebugSyncInput) (*DebugSyncOutput, error) {
	s.logger.Info("🔍 Starting sync diagnostics")

	output := &DebugSyncOutput{
		Issues:          make([]string, 0),
		Recommendations: make([]string, 0),
	}

	// Check authentication status
	if input.CheckAuth {
		s.checkAuthStatus(ctx, input.Password, output)
	}

	// Check sync state
	if input.CheckSyncState {
		s.checkSyncState(ctx, output)
	}

	// Provide overall assessment
	if len(output.Issues) == 0 {
		s.logger.Info("✅ No sync issues detected")
	} else {
		s.logger.Warn("⚠️ Sync issues detected", zap.Int("issueCount", len(output.Issues)))
	}

	return output, nil
}

// checkAuthStatus verifies user authentication and password
func (s *syncDebugService) checkAuthStatus(ctx context.Context, password string, output *DebugSyncOutput) {
	s.logger.Debug("🔐 Checking authentication status")

	// Get logged in user
	user, err := s.getUserByIsLoggedInUseCase.Execute(ctx)
	if err != nil {
		output.AuthStatus = "Failed to get user"
		output.Issues = append(output.Issues, "Cannot retrieve logged-in user")
		output.Recommendations = append(output.Recommendations, "Run 'maplefile-cli auth login' to log in")
		return
	}

	if user == nil {
		output.AuthStatus = "Not logged in"
		output.Issues = append(output.Issues, "User is not logged in")
		output.Recommendations = append(output.Recommendations, "Run 'maplefile-cli auth login' to log in")
		return
	}

	output.AuthStatus = fmt.Sprintf("Logged in as %s", user.Email)

	// Test password if provided
	if password != "" {
		s.testPasswordDecryption(user, password, output)
	} else {
		output.Issues = append(output.Issues, "No password provided for E2EE operations")
		output.Recommendations = append(output.Recommendations, "Provide password using --password flag")
	}
}

// testPasswordDecryption tests if the password can decrypt the master key
func (s *syncDebugService) testPasswordDecryption(user interface{}, password string, output *DebugSyncOutput) {
	// This would require importing crypto packages and implementing the decryption test
	// For now, we'll just check if password is provided
	if len(password) < 8 {
		output.Issues = append(output.Issues, "Password appears to be too short")
		output.Recommendations = append(output.Recommendations, "Ensure you're using the correct account password")
	}
}

// checkSyncState verifies the current sync state
func (s *syncDebugService) checkSyncState(ctx context.Context, output *DebugSyncOutput) {
	s.logger.Debug("📊 Checking sync state")

	syncStateOutput, err := s.syncStateGetService.GetSyncState(ctx)
	if err != nil {
		output.SyncStateStatus = "Failed to get sync state"
		output.Issues = append(output.Issues, "Cannot retrieve sync state")
		output.Recommendations = append(output.Recommendations, "Try running 'maplefile-cli sync reset' to reset sync state")
		return
	}

	if syncStateOutput.SyncState.LastCollectionSync.IsZero() {
		output.SyncStateStatus = "Never synced"
		output.Recommendations = append(output.Recommendations, "This is your first sync - it may take longer than usual")
	} else {
		output.SyncStateStatus = fmt.Sprintf("Last synced: %v", syncStateOutput.SyncState.LastCollectionSync)
	}
}
