// internal/repo/remotefile/delete.go
package remotefile

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Delete deletes a file from the cloud server
func (r *remoteFileRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	// Get server URL from configuration
	serverURL, err := r.configService.GetCloudProviderAddress(ctx)
	if err != nil {
		return errors.NewAppError("failed to get cloud provider address", err)
	}

	// Get access token
	accessToken, err := r.getAccessToken(ctx)
	if err != nil {
		return err
	}

	// Create HTTP request
	deleteURL := fmt.Sprintf("%s/maplefile/api/v1/files/%s", serverURL, id.Hex())
	req, err := http.NewRequestWithContext(ctx, "DELETE", deleteURL, nil)
	if err != nil {
		return errors.NewAppError("failed to create HTTP request", err)
	}

	req.Header.Set("Authorization", "JWT "+accessToken)

	// Execute the request
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return errors.NewAppError("failed to connect to server", err)
	}
	defer resp.Body.Close()

	// Check for error status codes
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return errors.NewAppError(fmt.Sprintf("server returned error status: %s, body: %s", resp.Status, string(body)), nil)
	}

	return nil
}
