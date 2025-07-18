// cloud/mapleapps-backend/internal/maplefile/service/dashboard/get_dashboard.go
package dashboard

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
	dom_dashboard "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/dashboard"
	file_service "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/service/file"
	uc_dashboard "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/dashboard"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type GetDashboardResponseDTO struct {
	Dashboard *dom_dashboard.DashboardData `json:"dashboard"`
	Success   bool                         `json:"success"`
	Message   string                       `json:"message"`
}

type GetDashboardService interface {
	Execute(ctx context.Context) (*GetDashboardResponseDTO, error)
}

type getDashboardServiceImpl struct {
	config                 *config.Configuration
	logger                 *zap.Logger
	getDashboardUseCase    uc_dashboard.GetDashboardUseCase
	listRecentFilesService file_service.ListRecentFilesService
}

func NewGetDashboardService(
	config *config.Configuration,
	logger *zap.Logger,
	getDashboardUseCase uc_dashboard.GetDashboardUseCase,
	listRecentFilesService file_service.ListRecentFilesService,
) GetDashboardService {
	logger = logger.Named("GetDashboardService")
	return &getDashboardServiceImpl{
		config:                 config,
		logger:                 logger,
		getDashboardUseCase:    getDashboardUseCase,
		listRecentFilesService: listRecentFilesService,
	}
}

func (svc *getDashboardServiceImpl) Execute(ctx context.Context) (*GetDashboardResponseDTO, error) {
	//
	// STEP 1: Get user ID from context
	//
	userID, ok := ctx.Value(constants.SessionFederatedUserID).(gocql.UUID)
	if !ok {
		svc.logger.Error("Failed getting user ID from context")
		return nil, httperror.NewForInternalServerErrorWithSingleField("message", "Authentication context error")
	}

	//
	// STEP 2: Get recent files using the working Recent Files Service
	//
	var recentFilesData []dom_dashboard.RecentFile
	recentFilesResp, err := svc.listRecentFilesService.Execute(ctx, nil, 5)
	if err != nil {
		svc.logger.Warn("Failed to get recent files for dashboard, using empty list",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		// Don't fail the entire dashboard for recent files
		recentFilesData = []dom_dashboard.RecentFile{}
	} else {
		// Convert recent files service response to dashboard domain model
		recentFilesData = svc.convertRecentFilesToDashboard(recentFilesResp.Files)
	}

	//
	// STEP 3: Execute dashboard use case with recent files data
	//
	dashboardResult, err := svc.getDashboardUseCase.Execute(ctx, userID, recentFilesData)
	if err != nil {
		svc.logger.Error("Failed to get dashboard data",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return nil, err
	}

	//
	// STEP 4: Build response
	//
	response := &GetDashboardResponseDTO{
		Dashboard: &dashboardResult.Dashboard,
		Success:   true,
		Message:   "Dashboard data retrieved successfully",
	}

	svc.logger.Info("Dashboard data retrieved successfully",
		zap.String("user_id", userID.String()),
		zap.Int("total_files", dashboardResult.Dashboard.Summary.TotalFiles),
		zap.Int("total_folders", dashboardResult.Dashboard.Summary.TotalFolders),
		zap.Int("storage_usage_percentage", dashboardResult.Dashboard.Summary.StorageUsagePercentage),
		zap.Int("recent_files_count", len(recentFilesData)))

	return response, nil
}

// convertRecentFilesToDashboard converts recent files service response to dashboard domain model
func (svc *getDashboardServiceImpl) convertRecentFilesToDashboard(recentFiles []file_service.RecentFileResponseDTO) []dom_dashboard.RecentFile {
	if len(recentFiles) == 0 {
		return []dom_dashboard.RecentFile{}
	}

	dashboardFiles := make([]dom_dashboard.RecentFile, len(recentFiles))
	for i, file := range recentFiles {
		// Extract placeholder file name from encrypted metadata
		fileName := svc.extractFileName(file.EncryptedMetadata)
		fileType := svc.determineFileType(fileName)
		uploadedTimeAgo := svc.formatTimeAgo(file.ModifiedAt)

		dashboardFiles[i] = dom_dashboard.RecentFile{
			FileName:          fileName,
			Uploaded:          uploadedTimeAgo,
			UploadedTimestamp: svc.parseTimestamp(file.ModifiedAt),
			Type:              fileType,
			Size:              svc.convertBytesToStorageAmount(file.EncryptedFileSizeInBytes),
		}
	}

	return dashboardFiles
}

// Helper methods for converting recent files data to dashboard format

func (svc *getDashboardServiceImpl) extractFileName(encryptedMetadata string) string {
	// In a real implementation, this would decrypt the metadata
	// For now, return a placeholder that indicates encrypted data
	if encryptedMetadata == "" {
		return "Unknown File"
	}
	// Could potentially extract some non-sensitive info or return a hash-based identifier
	maxLen := len(encryptedMetadata)
	if maxLen > 8 {
		maxLen = 8
	}
	return "File_" + encryptedMetadata[:maxLen]
}

func (svc *getDashboardServiceImpl) determineFileType(fileName string) string {
	// Simple file type determination based on extension
	if fileName == "" {
		return "Unknown"
	}

	// For placeholder names like "File_12345678", return "Document"
	if len(fileName) > 5 && fileName[:5] == "File_" {
		return "Document"
	}

	// This could be enhanced with actual extension parsing if we had real filenames
	return "Document"
}

func (svc *getDashboardServiceImpl) formatTimeAgo(timestampStr string) string {
	// Parse the timestamp string from the service response
	timestamp := svc.parseTimestamp(timestampStr)
	if timestamp.IsZero() {
		return "Unknown"
	}

	return svc.formatTimeAgoFromTime(timestamp)
}

func (svc *getDashboardServiceImpl) parseTimestamp(timestampStr string) time.Time {
	// Parse timestamp in format "2006-01-02T15:04:05Z07:00"
	timestamp, err := time.Parse("2006-01-02T15:04:05Z07:00", timestampStr)
	if err != nil {
		svc.logger.Warn("Failed to parse timestamp",
			zap.String("timestamp", timestampStr),
			zap.Error(err))
		return time.Time{}
	}
	return timestamp
}

func (svc *getDashboardServiceImpl) formatTimeAgoFromTime(timestamp time.Time) string {
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

func (svc *getDashboardServiceImpl) convertBytesToStorageAmount(bytes int64) dom_dashboard.StorageAmount {
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
