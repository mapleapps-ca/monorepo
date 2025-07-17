// cloud/mapleapps-backend/internal/maplefile/service/dashboard/get_dashboard.go
package dashboard

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
	dom_dashboard "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/dashboard"
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
	config              *config.Configuration
	logger              *zap.Logger
	getDashboardUseCase uc_dashboard.GetDashboardUseCase
}

func NewGetDashboardService(
	config *config.Configuration,
	logger *zap.Logger,
	getDashboardUseCase uc_dashboard.GetDashboardUseCase,
) GetDashboardService {
	logger = logger.Named("GetDashboardService")
	return &getDashboardServiceImpl{
		config:              config,
		logger:              logger,
		getDashboardUseCase: getDashboardUseCase,
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
	// STEP 2: Execute dashboard use case
	//
	dashboardResult, err := svc.getDashboardUseCase.Execute(ctx, userID)
	if err != nil {
		svc.logger.Error("Failed to get dashboard data",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return nil, err
	}

	//
	// STEP 3: Build response
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
		zap.Int("storage_usage_percentage", dashboardResult.Dashboard.Summary.StorageUsagePercentage))

	return response, nil
}
