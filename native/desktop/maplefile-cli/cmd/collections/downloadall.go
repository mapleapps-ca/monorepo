// monorepo/native/desktop/maplefile-cli/cmd/collections/downloadall.go
package collections

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/collectionsyncer"
)

// downloadAllCollectionsCmd creates a command for downloading all collections from the cloud
func downloadAllCollectionsCmd(
	downloadService collectionsyncer.DownloadService,
	logger *zap.Logger,
) *cobra.Command {
	var verbose bool
	var timeout int

	var cmd = &cobra.Command{
		Use:   "downloadall",
		Short: "Download all collections from the cloud",
		Long: `
Download all collections from the cloud to local storage.

This command synchronizes all your remote collections to your local device,
allowing you to work with them offline and providing faster access.

Examples:
  # Download all collections
  maplefile-cli collections downloadall

  # Download all collections with verbose output
  maplefile-cli collections downloadall --verbose

  # Download all collections with a longer timeout
  maplefile-cli collections downloadall --timeout 120
`,
		Run: func(cmd *cobra.Command, args []string) {
			// Create a context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
			defer cancel()

			if verbose {
				fmt.Println("Starting download of all collections from the cloud...")
				fmt.Printf("Using timeout of %d seconds\n", timeout)
			} else {
				fmt.Println("Downloading collections...")
			}

			// Start timer to measure download duration
			startTime := time.Now()

			// Download all collections
			count, err := downloadService.DownloadAll(ctx)
			if err != nil {
				fmt.Printf("🐞 Error downloading collections: %v\n", err)
				return
			}

			// Calculate download duration
			duration := time.Since(startTime)

			// Display success message
			fmt.Println("\n✅ Download completed successfully!")
			fmt.Printf("Downloaded %d collections\n", count)

			if verbose {
				fmt.Printf("Download took %s\n", duration.Round(time.Millisecond))
			}
		},
	}

	// Define command flags
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	cmd.Flags().IntVarP(&timeout, "timeout", "t", 60, "Timeout in seconds for the download operation")

	return cmd
}
