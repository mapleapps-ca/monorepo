// native/desktop/maplefile-cli/internal/domain/medto/interface.go
package medto

import (
	"context"
)

// MeDTORepository defines the interface for interacting with the cloud service
// to manage user profile (Me) data. These DTOs represent user profile data
// exchanged between the local device and the cloud server.
type MeDTORepository interface {
	// GetMeFromCloud fetches the current user's profile from the cloud service.
	// It returns the MeResponseDTO if successful, or an error if the operation fails.
	GetMeFromCloud(ctx context.Context) (*MeResponseDTO, error)

	// UpdateMeInCloud updates the current user's profile in the cloud service.
	// It takes an UpdateMeRequestDTO and returns the updated MeResponseDTO if successful,
	// or an error if the operation fails.
	UpdateMeInCloud(ctx context.Context, request *UpdateMeRequestDTO) (*MeResponseDTO, error)
}
