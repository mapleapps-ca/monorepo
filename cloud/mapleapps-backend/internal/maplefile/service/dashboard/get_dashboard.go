// cloud/mapleapps-backend/internal/maplefile/service/dashboard/get_dashboard.go
package dashboard

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
	dom_feduser "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/federateduser"
	uc_feduser "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/usecase/federateduser"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/storagedailyusage"
	file_service "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/service/file"
	uc_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/collection"
	uc_filemetadata "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/filemetadata"
	uc_storagedailyusage "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/storagedailyusage"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type GetDashboardService interface {
	Execute(ctx context.Context) (*GetDashboardResponseDTO, error)
}

type getDashboardServiceImpl struct {
	config                      *config.Configuration
	logger                      *zap.Logger
	listRecentFilesService      file_service.ListRecentFilesService
	federatedUserGetByIDUseCase uc_feduser.FederatedUserGetByIDUseCase
	countUserFilesUseCase       uc_filemetadata.CountUserFilesUseCase
	countUserCollectionsUseCase uc_collection.CountUserCollectionsUseCase
	getStorageTrendUseCase      uc_storagedailyusage.GetStorageDailyUsageTrendUseCase
}

func NewGetDashboardService(
	config *config.Configuration,
	logger *zap.Logger,
	listRecentFilesService file_service.ListRecentFilesService,
	federatedUserGetByIDUseCase uc_feduser.FederatedUserGetByIDUseCase,
	countUserFilesUseCase uc_filemetadata.CountUserFilesUseCase,
	countUserCollectionsUseCase uc_collection.CountUserCollectionsUseCase,
	getStorageTrendUseCase uc_storagedailyusage.GetStorageDailyUsageTrendUseCase,
) GetDashboardService {
	logger = logger.Named("GetDashboardService")
	return &getDashboardServiceImpl{
		config:                      config,
		logger:                      logger,
		listRecentFilesService:      listRecentFilesService,
		federatedUserGetByIDUseCase: federatedUserGetByIDUseCase,
		countUserFilesUseCase:       countUserFilesUseCase,
		countUserCollectionsUseCase: countUserCollectionsUseCase,
		getStorageTrendUseCase:      getStorageTrendUseCase,
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
	// STEP 2: Validation
	//
	e := make(map[string]string)
	if userID.String() == "" {
		e["user_id"] = "User ID is required"
	}
	if len(e) != 0 {
		svc.logger.Warn("Failed validating get dashboard",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 3: Get user information for storage data
	//
	user, err := svc.federatedUserGetByIDUseCase.Execute(ctx, userID)
	if err != nil {
		svc.logger.Error("Failed to get user for dashboard",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return nil, err
	}

	if user == nil {
		svc.logger.Warn("User not found for dashboard",
			zap.String("user_id", userID.String()))
		return nil, httperror.NewForNotFoundWithSingleField("user_id", "User not found")
	}

	//
	// STEP 4: Get file count
	//
	fileCountResp, err := svc.countUserFilesUseCase.Execute(ctx, userID)
	if err != nil {
		svc.logger.Error("Failed to count user files for dashboard",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return nil, err
	}

	//
	// STEP 5: Get collection count
	//
	collectionCountResp, err := svc.countUserCollectionsUseCase.Execute(ctx, userID)
	if err != nil {
		svc.logger.Error("Failed to count user collections for dashboard",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return nil, err
	}

	// Debug logging for collection count issue
	svc.logger.Debug("Collection count debug info",
		zap.String("user_id", userID.String()),
		zap.Int("total_collections_returned", collectionCountResp.TotalCollections))

	//
	// STEP 6: Get storage usage trend (last 7 days)
	//
	trendReq := &uc_storagedailyusage.GetStorageDailyUsageTrendRequest{
		UserID:      userID,
		TrendPeriod: "7days",
	}

	storageTrend, err := svc.getStorageTrendUseCase.Execute(ctx, trendReq)
	if err != nil {
		svc.logger.Warn("Failed to get storage trend for dashboard, using empty trend",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		// Don't fail the entire dashboard for trend data
		storageTrend = nil
	}

	//
	// STEP 7: Get recent files using the working Recent Files Service
	//
	var recentFiles []file_service.RecentFileResponseDTO
	recentFilesResp, err := svc.listRecentFilesService.Execute(ctx, nil, 5)
	if err != nil {
		svc.logger.Warn("Failed to get recent files for dashboard, using empty list",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		// Don't fail the entire dashboard for recent files
		recentFiles = []file_service.RecentFileResponseDTO{}
	} else {
		recentFiles = recentFilesResp.Files
	}

	//
	// STEP 8: Build dashboard response
	//
	dashboard := &DashboardDataDTO{
		Summary:           svc.buildSummary(user, fileCountResp.TotalFiles, collectionCountResp.TotalCollections),
		StorageUsageTrend: svc.buildStorageUsageTrend(storageTrend),
		RecentFiles:       recentFiles,
	}

	response := &GetDashboardResponseDTO{
		Dashboard: dashboard,
		Success:   true,
		Message:   "Dashboard data retrieved successfully",
	}

	svc.logger.Info("Dashboard data retrieved successfully",
		zap.String("user_id", userID.String()),
		zap.Int("total_files", fileCountResp.TotalFiles),
		zap.Int("total_folders", collectionCountResp.TotalCollections),
		zap.Int("recent_files_count", len(recentFiles)))

	return response, nil
}

func (svc *getDashboardServiceImpl) buildSummary(user *dom_feduser.FederatedUser, totalFiles, totalFolders int) SummaryDTO {
	// Convert storage used to human-readable format
	storageUsed := svc.convertBytesToStorageAmount(user.StorageUsedBytes)
	storageLimit := svc.convertBytesToStorageAmount(user.StorageLimitBytes)

	// Calculate storage percentage manually to fix potential issues
	storagePercentage := 0
	if user.StorageLimitBytes > 0 {
		percentage := (float64(user.StorageUsedBytes) / float64(user.StorageLimitBytes)) * 100
		storagePercentage = int(percentage)
	}

	// Debug logging for storage calculation
	svc.logger.Debug("Storage calculation debug",
		zap.Int64("storage_used_bytes", user.StorageUsedBytes),
		zap.Int64("storage_limit_bytes", user.StorageLimitBytes),
		zap.Int("calculated_percentage", storagePercentage),
		zap.Float64("user_method_percentage", user.GetStorageUsagePercentage()))

	return SummaryDTO{
		TotalFiles:             totalFiles,
		TotalFolders:           totalFolders,
		StorageUsed:            storageUsed,
		StorageLimit:           storageLimit,
		StorageUsagePercentage: storagePercentage, // Use our manual calculation
	}
}

func (svc *getDashboardServiceImpl) buildStorageUsageTrend(trend *storagedailyusage.StorageUsageTrend) StorageUsageTrendDTO {
	if trend == nil || len(trend.DailyUsages) == 0 {
		return StorageUsageTrendDTO{
			Period:     "Last 7 days",
			DataPoints: []DataPointDTO{},
		}
	}

	dataPoints := make([]DataPointDTO, len(trend.DailyUsages))
	for i, daily := range trend.DailyUsages {
		dataPoints[i] = DataPointDTO{
			Date:  daily.UsageDay.Format("2006-01-02"),
			Usage: svc.convertBytesToStorageAmount(daily.TotalBytes),
		}
	}

	return StorageUsageTrendDTO{
		Period:     "Last 7 days",
		DataPoints: dataPoints,
	}
}

func (svc *getDashboardServiceImpl) convertBytesToStorageAmount(bytes int64) StorageAmountDTO {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return StorageAmountDTO{
			Value: float64(bytes) / TB,
			Unit:  "TB",
		}
	case bytes >= GB:
		return StorageAmountDTO{
			Value: float64(bytes) / GB,
			Unit:  "GB",
		}
	case bytes >= MB:
		return StorageAmountDTO{
			Value: float64(bytes) / MB,
			Unit:  "MB",
		}
	case bytes >= KB:
		return StorageAmountDTO{
			Value: float64(bytes) / KB,
			Unit:  "KB",
		}
	default:
		return StorageAmountDTO{
			Value: float64(bytes),
			Unit:  "B",
		}
	}
}
