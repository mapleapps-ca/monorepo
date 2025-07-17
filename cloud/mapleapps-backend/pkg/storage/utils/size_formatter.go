// monorepo/cloud/mapleapps-backend/pkg/storage/utils/size_formatter.go
package utils

import (
	"fmt"
	"math"
)

// StorageSizeUnit represents different storage size units
type StorageSizeUnit string

const (
	UnitBytes     StorageSizeUnit = "B"
	UnitKilobytes StorageSizeUnit = "KB"
	UnitMegabytes StorageSizeUnit = "MB"
	UnitGigabytes StorageSizeUnit = "GB"
	UnitTerabytes StorageSizeUnit = "TB"
	UnitPetabytes StorageSizeUnit = "PB"
)

// FormattedSize represents a storage size with value and unit
type FormattedSize struct {
	Value float64         `json:"value"`
	Unit  StorageSizeUnit `json:"unit"`
	Raw   int64           `json:"raw_bytes"`
}

// String returns a human-readable string representation
func (fs FormattedSize) String() string {
	if fs.Value == math.Trunc(fs.Value) {
		return fmt.Sprintf("%.0f %s", fs.Value, fs.Unit)
	}
	return fmt.Sprintf("%.2f %s", fs.Value, fs.Unit)
}

// FormatBytes converts bytes to a human-readable format
func FormatBytes(bytes int64) FormattedSize {
	if bytes == 0 {
		return FormattedSize{Value: 0, Unit: UnitBytes, Raw: 0}
	}

	const unit = 1024
	if bytes < unit {
		return FormattedSize{
			Value: float64(bytes),
			Unit:  UnitBytes,
			Raw:   bytes,
		}
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []StorageSizeUnit{UnitKilobytes, UnitMegabytes, UnitGigabytes, UnitTerabytes, UnitPetabytes}

	return FormattedSize{
		Value: math.Round(float64(bytes)/float64(div)*100) / 100,
		Unit:  units[exp],
		Raw:   bytes,
	}
}

// FormatBytesWithPrecision converts bytes to human-readable format with specified decimal places
func FormatBytesWithPrecision(bytes int64, precision int) FormattedSize {
	formatted := FormatBytes(bytes)

	// Round to specified precision
	multiplier := math.Pow(10, float64(precision))
	formatted.Value = math.Round(formatted.Value*multiplier) / multiplier

	return formatted
}

// Enhanced response types with formatted sizes
type StorageSizeResponseFormatted struct {
	TotalSizeBytes     int64         `json:"total_size_bytes"`
	TotalSizeFormatted FormattedSize `json:"total_size_formatted"`
}

type StorageSizeBreakdownResponseFormatted struct {
	OwnedSizeBytes               int64                    `json:"owned_size_bytes"`
	OwnedSizeFormatted           FormattedSize            `json:"owned_size_formatted"`
	SharedSizeBytes              int64                    `json:"shared_size_bytes"`
	SharedSizeFormatted          FormattedSize            `json:"shared_size_formatted"`
	TotalSizeBytes               int64                    `json:"total_size_bytes"`
	TotalSizeFormatted           FormattedSize            `json:"total_size_formatted"`
	CollectionBreakdownBytes     map[string]int64         `json:"collection_breakdown_bytes"`
	CollectionBreakdownFormatted map[string]FormattedSize `json:"collection_breakdown_formatted"`
	OwnedCollectionsCount        int                      `json:"owned_collections_count"`
	SharedCollectionsCount       int                      `json:"shared_collections_count"`
}

// Example usage and outputs:
/*
FormatBytes(1024) -> {Value: 1, Unit: "KB", Raw: 1024} -> "1 KB"
FormatBytes(1536) -> {Value: 1.5, Unit: "KB", Raw: 1536} -> "1.50 KB"
FormatBytes(1073741824) -> {Value: 1, Unit: "GB", Raw: 1073741824} -> "1 GB"
FormatBytes(2684354560) -> {Value: 2.5, Unit: "GB", Raw: 2684354560} -> "2.50 GB"

Example formatted response:
{
  "total_size_bytes": 2684354560,
  "total_size_formatted": {
    "value": 2.5,
    "unit": "GB",
    "raw_bytes": 2684354560
  }
}
*/
