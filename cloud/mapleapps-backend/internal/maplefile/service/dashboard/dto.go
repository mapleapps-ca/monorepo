// cloud/mapleapps-backend/internal/maplefile/service/dashboard/dto.go
package dashboard

import (
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/service/file"
)

// GetDashboardResponseDTO represents the complete dashboard response
type GetDashboardResponseDTO struct {
	Dashboard *DashboardDataDTO `json:"dashboard"`
	Success   bool              `json:"success"`
	Message   string            `json:"message"`
}

// DashboardDataDTO contains all the dashboard information
type DashboardDataDTO struct {
	Summary           SummaryDTO                   `json:"summary"`
	StorageUsageTrend StorageUsageTrendDTO         `json:"storage_usage_trend"`
	RecentFiles       []file.RecentFileResponseDTO `json:"recent_files"`
}

// SummaryDTO contains the main dashboard statistics
type SummaryDTO struct {
	TotalFiles             int              `json:"total_files"`
	TotalFolders           int              `json:"total_folders"`
	StorageUsed            StorageAmountDTO `json:"storage_used"`
	StorageLimit           StorageAmountDTO `json:"storage_limit"`
	StorageUsagePercentage int              `json:"storage_usage_percentage"`
}

// StorageAmountDTO represents a storage value with its unit
type StorageAmountDTO struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

// StorageUsageTrendDTO contains the trend chart data
type StorageUsageTrendDTO struct {
	Period     string         `json:"period"`
	DataPoints []DataPointDTO `json:"data_points"`
}

// DataPointDTO represents a single point in the storage usage trend
type DataPointDTO struct {
	Date  string           `json:"date"`
	Usage StorageAmountDTO `json:"usage"`
}
