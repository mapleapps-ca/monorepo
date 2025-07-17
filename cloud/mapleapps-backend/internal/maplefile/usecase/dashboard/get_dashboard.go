// cloud/mapleapps-backend/internal/maplefile/usecase/dashboard/get_dashboard.go
package dashboard

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_feduser "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/federateduser"
	uc_feduser "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/usecase/federateduser"
	dom_dashboard "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/dashboard"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/storagedailyusage"
	uc_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/collection"
	uc_filemetadata "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/filemetadata"
	uc_storagedailyusage "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/storagedailyusage"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type GetDashboardUseCase interface {
	Execute(ctx context.Context, userID gocql.UUID) (*dom_dashboard.Dashboard, error)
}

type getDashboardUseCaseImpl struct {
	config                      *config.Configuration
	logger                      *zap.Logger
	federatedUserGetByIDUseCase uc_feduser.FederatedUserGetByIDUseCase
	countUserFilesUseCase       uc_filemetadata.CountUserFilesUseCase
	countUserCollectionsUseCase uc_collection.CountUserCollectionsUseCase
	getStorageTrendUseCase      uc_storagedailyusage.GetStorageDailyUsageTrendUseCase
	listRecentFilesUseCase      uc_filemetadata.ListRecentFilesUseCase
}

func NewGetDashboardUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	federatedUserGetByIDUseCase uc_feduser.FederatedUserGetByIDUseCase,
	countUserFilesUseCase uc_filemetadata.CountUserFilesUseCase,
	countUserCollectionsUseCase uc_collection.CountUserCollectionsUseCase,
	getStorageTrendUseCase uc_storagedailyusage.GetStorageDailyUsageTrendUseCase,
	listRecentFilesUseCase uc_filemetadata.ListRecentFilesUseCase,
) GetDashboardUseCase {
	logger = logger.Named("GetDashboardUseCase")
	return &getDashboardUseCaseImpl{
		config:                      config,
		logger:                      logger,
		federatedUserGetByIDUseCase: federatedUserGetByIDUseCase,
		countUserFilesUseCase:       countUserFilesUseCase,
		countUserCollectionsUseCase: countUserCollectionsUseCase,
		getStorageTrendUseCase:      getStorageTrendUseCase,
		listRecentFilesUseCase:      listRecentFilesUseCase,
	}
}

func (uc *getDashboardUseCaseImpl) Execute(ctx context.Context, userID gocql.UUID) (*dom_dashboard.Dashboard, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if userID.String() == "" {
		e["user_id"] = "User ID is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating get dashboard",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Get user information for storage data.
	//

	user, err := uc.federatedUserGetByIDUseCase.Execute(ctx, userID)
	if err != nil {
		uc.logger.Error("Failed to get user for dashboard",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return nil, err
	}

	if user == nil {
		uc.logger.Warn("User not found for dashboard",
			zap.String("user_id", userID.String()))
		return nil, httperror.NewForNotFoundWithSingleField("user_id", "User not found")
	}

	//
	// STEP 3: Get file count.
	//

	fileCountResp, err := uc.countUserFilesUseCase.Execute(ctx, userID)
	if err != nil {
		uc.logger.Error("Failed to count user files for dashboard",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return nil, err
	}

	//
	// STEP 4: Get collection count.
	//

	collectionCountResp, err := uc.countUserCollectionsUseCase.Execute(ctx, userID)
	if err != nil {
		uc.logger.Error("Failed to count user collections for dashboard",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return nil, err
	}

	//
	// STEP 5: Get storage usage trend (last 7 days).
	//

	trendReq := &uc_storagedailyusage.GetStorageDailyUsageTrendRequest{
		UserID:      userID,
		TrendPeriod: "7days",
	}

	storageTrend, err := uc.getStorageTrendUseCase.Execute(ctx, trendReq)
	if err != nil {
		uc.logger.Warn("Failed to get storage trend for dashboard, using empty trend",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		// Don't fail the entire dashboard for trend data
		storageTrend = nil
	}

	//
	// STEP 6: Get recent files (limit to 5 for dashboard).
	//

	recentFilesResp, err := uc.listRecentFilesUseCase.Execute(ctx, userID, nil, 5)
	if err != nil {
		uc.logger.Warn("Failed to get recent files for dashboard, using empty list",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		// Don't fail the entire dashboard for recent files
		recentFilesResp = nil
	}

	//
	// STEP 7: Build dashboard response.
	//

	dashboard := &dom_dashboard.Dashboard{
		Dashboard: dom_dashboard.DashboardData{
			Summary:           uc.buildSummary(user, fileCountResp.TotalFiles, collectionCountResp.TotalCollections),
			StorageUsageTrend: uc.buildStorageUsageTrend(storageTrend),
			RecentFiles:       uc.buildRecentFiles(recentFilesResp),
		},
	}

	uc.logger.Info("Successfully built dashboard data",
		zap.String("user_id", userID.String()),
		zap.Int("total_files", fileCountResp.TotalFiles),
		zap.Int("total_collections", collectionCountResp.TotalCollections),
		zap.Float64("storage_usage_percentage", user.GetStorageUsagePercentage()))

	return dashboard, nil
}

func (uc *getDashboardUseCaseImpl) buildSummary(user *dom_feduser.FederatedUser, totalFiles, totalFolders int) dom_dashboard.Summary {
	// Convert storage used to human-readable format
	storageUsed := uc.convertBytesToStorageAmount(user.StorageUsedBytes)
	storageLimit := uc.convertBytesToStorageAmount(user.StorageLimitBytes)

	return dom_dashboard.Summary{
		TotalFiles:             totalFiles,
		TotalFolders:           totalFolders,
		StorageUsed:            storageUsed,
		StorageLimit:           storageLimit,
		StorageUsagePercentage: int(user.GetStorageUsagePercentage()),
	}
}

func (uc *getDashboardUseCaseImpl) buildStorageUsageTrend(trend *storagedailyusage.StorageUsageTrend) dom_dashboard.StorageUsageTrend {
	if trend == nil || len(trend.DailyUsages) == 0 {
		return dom_dashboard.StorageUsageTrend{
			Period:     "Last 7 days",
			DataPoints: []dom_dashboard.DataPoint{},
		}
	}

	dataPoints := make([]dom_dashboard.DataPoint, len(trend.DailyUsages))
	for i, daily := range trend.DailyUsages {
		dataPoints[i] = dom_dashboard.DataPoint{
			Date:  daily.UsageDay.Format("2006-01-02"),
			Usage: uc.convertBytesToStorageAmount(daily.TotalBytes),
		}
	}

	return dom_dashboard.StorageUsageTrend{
		Period:     "Last 7 days",
		DataPoints: dataPoints,
	}
}

func (uc *getDashboardUseCaseImpl) buildRecentFiles(recentFilesResp *dom_file.RecentFilesResponse) []dom_dashboard.RecentFile {
	if recentFilesResp == nil || len(recentFilesResp.Files) == 0 {
		return []dom_dashboard.RecentFile{}
	}

	recentFiles := make([]dom_dashboard.RecentFile, len(recentFilesResp.Files))
	for i, file := range recentFilesResp.Files {
		// Extract file name from encrypted metadata (would need decryption in real scenario)
		fileName := uc.extractFileName(file.EncryptedMetadata)
		fileType := uc.determineFileType(fileName)
		uploadedTimeAgo := uc.formatTimeAgo(file.ModifiedAt)

		recentFiles[i] = dom_dashboard.RecentFile{
			FileName:          fileName,
			Uploaded:          uploadedTimeAgo,
			UploadedTimestamp: file.ModifiedAt,
			Type:              fileType,
			Size:              uc.convertBytesToStorageAmount(file.EncryptedFileSizeInBytes),
		}
	}

	return recentFiles
}

func (uc *getDashboardUseCaseImpl) convertBytesToStorageAmount(bytes int64) dom_dashboard.StorageAmount {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return dom_dashboard.StorageAmount{
			Value: float64(bytes) / TB,
			Unit:  "TB",
		}
	case bytes >= GB:
		return dom_dashboard.StorageAmount{
			Value: float64(bytes) / GB,
			Unit:  "GB",
		}
	case bytes >= MB:
		return dom_dashboard.StorageAmount{
			Value: float64(bytes) / MB,
			Unit:  "MB",
		}
	case bytes >= KB:
		return dom_dashboard.StorageAmount{
			Value: float64(bytes) / KB,
			Unit:  "KB",
		}
	default:
		return dom_dashboard.StorageAmount{
			Value: float64(bytes),
			Unit:  "B",
		}
	}
}

func (uc *getDashboardUseCaseImpl) extractFileName(encryptedMetadata string) string {
	// In a real implementation, this would decrypt the metadata
	// For now, return a placeholder that indicates encrypted data
	if encryptedMetadata == "" {
		return "Unknown File"
	}
	// Could potentially extract some non-sensitive info or return a hash-based identifier
	return fmt.Sprintf("File_%s", encryptedMetadata[:min(8, len(encryptedMetadata))])
}

func (uc *getDashboardUseCaseImpl) determineFileType(fileName string) string {
	// Simple file type determination based on extension
	if fileName == "" {
		return "Unknown"
	}

	// Extract extension
	parts := strings.Split(fileName, ".")
	if len(parts) < 2 {
		return "Document"
	}

	ext := strings.ToLower(parts[len(parts)-1])
	switch ext {
	case "jpg", "jpeg", "png", "gif", "bmp", "svg":
		return "Image"
	case "mp4", "avi", "mov", "wmv", "flv", "webm":
		return "Video"
	case "mp3", "wav", "flac", "aac", "ogg":
		return "Audio"
	case "pdf":
		return "PDF"
	case "doc", "docx":
		return "Word Document"
	case "xls", "xlsx":
		return "Spreadsheet"
	case "ppt", "pptx":
		return "Presentation"
	case "txt":
		return "Text"
	case "zip", "rar", "7z", "tar", "gz":
		return "Archive"
	default:
		return "Document"
	}
}

func (uc *getDashboardUseCaseImpl) formatTimeAgo(timestamp time.Time) string {
	now := time.Now()
	diff := now.Sub(timestamp)

	switch {
	case diff < time.Minute:
		return "Just now"
	case diff < time.Hour:
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	default:
		return timestamp.Format("Jan 2, 2006")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
