// native/desktop/maplefile-cli/internal/domain/recovery/dto_interface.go
package recovery

import (
	"context"
)

// RecoveryDTORepository defines the interface for recovery cloud operations
// This interface follows the naming convention where DTO repositories make API calls
// to/from the cloud service, as opposed to local database operations
type RecoveryDTORepository interface {
	// Core recovery flow operations with cloud service

	// InitiateRecoveryFromCloud starts the recovery process with the cloud service
	// This corresponds to POST /iam/api/v1/recovery/initiate
	InitiateRecoveryFromCloud(ctx context.Context, request *RecoveryInitiateRequest) (*RecoveryInitiateResponse, error)

	// VerifyRecoveryFromCloud verifies the recovery challenge with the cloud service
	// This corresponds to POST /iam/api/v1/recovery/verify
	VerifyRecoveryFromCloud(ctx context.Context, request *RecoveryVerifyRequest) (*RecoveryVerifyResponse, error)

	// CompleteRecoveryFromCloud completes the recovery process with the cloud service
	// This corresponds to POST /iam/api/v1/recovery/complete
	CompleteRecoveryFromCloud(ctx context.Context, request *RecoveryCompleteRequest) (*RecoveryCompleteResponse, error)

	// Additional cloud operations (if needed in the future)

	// GetRecoveryStatusFromCloud gets the current status of a recovery session from cloud
	// GetRecoveryStatusFromCloud(ctx context.Context, sessionID string) (*RecoveryStatusResponse, error)

	// CancelRecoveryInCloud cancels an active recovery session in the cloud
	// CancelRecoveryInCloud(ctx context.Context, sessionID string) error
}
