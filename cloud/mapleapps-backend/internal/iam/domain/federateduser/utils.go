// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/federateduser/utils.go
package federateduser

import "time"

// CanUpload checks if user has enough storage quota to upload a file
func (u *FederatedUser) CanUpload(fileSize int64) bool {
	return u.StorageUsedBytes+fileSize <= u.StorageLimitBytes
}

// UpgradePlan upgrades the user's plan and adjusts storage limit
func (u *FederatedUser) UpgradePlan(newPlan string) {
	u.UserPlan = newPlan
	u.StorageLimitBytes = GetStorageLimitForFederatedUserPlan(newPlan)
	u.ModifiedAt = time.Now().UTC()
}

// GetStorageUsagePercentage returns the percentage of storage used
func (u *FederatedUser) GetStorageUsagePercentage() float64 {
	if u.StorageLimitBytes == 0 {
		if u.UserPlan != "" {
			return float64(GetStorageLimitForFederatedUserPlan(u.UserPlan))
		}
		return 0
	}
	return (float64(u.StorageUsedBytes) / float64(u.StorageLimitBytes)) * 100
}

// GetRemainingStorage returns the amount of storage left in bytes
func (u *FederatedUser) GetRemainingStorage() int64 {
	remaining := u.StorageLimitBytes - u.StorageUsedBytes
	if remaining < 0 {
		return 0
	}
	return remaining
}
