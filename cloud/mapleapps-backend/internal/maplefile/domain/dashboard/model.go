// cloud/mapleapps-backend/internal/maplefile/domain/dashboard/model.go
package dashboard

import (
	"time"
)

// Dashboard represents the main dashboard data structure
type Dashboard struct {
	Dashboard DashboardData `json:"dashboard"`
}

// DashboardData contains all the dashboard information
type DashboardData struct {
	Summary           Summary           `json:"summary"`
	StorageUsageTrend StorageUsageTrend `json:"storageUsageTrend"`
	RecentFiles       []RecentFile      `json:"recentFiles"`
}

// Summary contains the main dashboard statistics
type Summary struct {
	TotalFiles             int           `json:"totalFiles"`
	TotalFolders           int           `json:"totalFolders"`
	StorageUsed            StorageAmount `json:"storageUsed"`
	StorageLimit           StorageAmount `json:"storageLimit"`
	StorageUsagePercentage int           `json:"storageUsagePercentage"`
}

// StorageAmount represents a storage value with its unit
type StorageAmount struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

// StorageUsageTrend contains the trend chart data
type StorageUsageTrend struct {
	Period     string      `json:"period"`
	DataPoints []DataPoint `json:"dataPoints"`
}

// DataPoint represents a single point in the storage usage trend
type DataPoint struct {
	Date  string        `json:"date"`
	Usage StorageAmount `json:"usage"`
}

// RecentFile represents a file in the recent files list
type RecentFile struct {
	FileName          string        `json:"fileName"`
	Uploaded          string        `json:"uploaded"`
	UploadedTimestamp time.Time     `json:"uploadedTimestamp"`
	Type              string        `json:"type"`
	Size              StorageAmount `json:"size"`
}
