// monorepo/native/desktop/maplefile-cli/cmd/collections/create.go
package collections

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/collectionsyncer"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/remotecollection"
)

// createRootCollectionCmd creates a command for creating a root collection
func createRootCollectionCmd(
	remoteCollectionService remotecollection.CreateService,
	downloadService collectionsyncer.DownloadService,
	logger *zap.Logger,
) *cobra.Command {
	var name, collectionType string

	var cmd = &cobra.Command{
		Use:   "create",
		Short: "Create a new root collection",
		Long: `
Create a new root collection to store files.

A collection is an encrypted container for your files. Root collections appear
at the top level of your collection hierarchy.

Examples:
  # Create a folder collection
  maplefile-cli collections create --name "My Documents" --type folder

  # Create an album collection
  maplefile-cli collections create --name "Photo Album" --type album
`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			// Validate required fields
			if name == "" {
				fmt.Println("Error: Collection name is required.")
				fmt.Println("Use --name flag to specify the name of your collection.")
				return
			}

			// If type is not specified, default to folder
			if collectionType == "" {
				collectionType = "folder"
			}

			// Ensure valid collection type
			if collectionType != "folder" && collectionType != "album" {
				fmt.Printf("Error: Invalid collection type: %s\n", collectionType)
				fmt.Println("Collection type must be either 'folder' or 'album'.")
				return
			}

			// Step 1: Create the collection remotely
			remoteInput := remotecollection.CreateInput{
				Name:           name,
				CollectionType: collectionType,
			}

			logger.Debug("Creating remote collection",
				zap.String("name", name),
				zap.String("type", collectionType))

			remoteOutput, err := remoteCollectionService.Create(ctx, remoteInput)
			if err != nil {
				fmt.Printf("🐞 Error creating remote collection: %v\n", err)
				return
			}

			// Step 2: Download the remote collection to local storage
			remoteID := remoteOutput.Collection.ID.Hex()
			logger.Debug("Downloading remote collection to local storage",
				zap.String("remoteID", remoteID))

			downloadOutput, err := downloadService.Download(ctx, remoteID)
			if err != nil {
				fmt.Printf("🐞 Warning: Created remote collection, but failed to download to local storage: %v\n", err)
				// Continue execution, as the remote collection was successfully created
			}

			// Display success message
			fmt.Println("\n✅ Collection created successfully!")
			fmt.Printf("Collection ID: %s\n", remoteOutput.Collection.ID.Hex())
			fmt.Printf("Collection Type: %s\n", remoteOutput.Collection.Type)
			fmt.Printf("Created At: %s\n", remoteOutput.Collection.CreatedAt.Format("2006-01-02 15:04:05"))

			if downloadOutput != nil && downloadOutput.Collection != nil {
				fmt.Println("Collection synchronized to local storage.")
			}
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&name, "name", "n", "", "Name of the collection (required)")
	cmd.Flags().StringVarP(&collectionType, "type", "t", "folder", "Type of collection ('folder' or 'album')")

	// Mark required flags
	cmd.MarkFlagRequired("name")

	return cmd
}

// createSubCollectionCmd creates a command for creating a sub-collection
func createSubCollectionCmd(
	remoteCollectionService remotecollection.CreateService,
	downloadService collectionsyncer.DownloadService,
	logger *zap.Logger,
) *cobra.Command {
	var name, collectionType, parentID string

	var cmd = &cobra.Command{
		Use:   "create-sub",
		Short: "Create a new sub-collection",
		Long: `
Create a new sub-collection within an existing collection.

Sub-collections allow you to organize your files hierarchically.
You must specify the parent collection ID where the new sub-collection will be created.

Examples:
  # Create a folder sub-collection
  maplefile-cli collections create-sub --name "Project Documents" --type folder --parent 507f1f77bcf86cd799439011

  # Create an album sub-collection
  maplefile-cli collections create-sub --name "Vacation Photos" --type album --parent 507f1f77bcf86cd799439011
`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			// Validate required fields
			if name == "" {
				fmt.Println("🐞 Error: Collection name is required.")
				fmt.Println("Use --name flag to specify the name of your collection.")
				return
			}

			if parentID == "" {
				fmt.Println("🐞 Error: Parent collection ID is required.")
				fmt.Println("Use --parent flag to specify the ID of the parent collection.")
				return
			}

			// If type is not specified, default to folder
			if collectionType == "" {
				collectionType = "folder"
			}

			// Ensure valid collection type
			if collectionType != "folder" && collectionType != "album" {
				fmt.Printf("🐞 Error: Invalid collection type: %s\n", collectionType)
				fmt.Println("Collection type must be either 'folder' or 'album'.")
				return
			}

			// Validate parentID format
			_, err := primitive.ObjectIDFromHex(parentID)
			if err != nil {
				fmt.Printf("🐞 Error: Invalid parent ID format: %v\n", err)
				return
			}

			// Step 1: Create the sub-collection remotely
			remoteInput := remotecollection.CreateInput{
				Name:           name,
				CollectionType: collectionType,
				ParentID:       parentID,
			}

			logger.Debug("Creating remote sub-collection",
				zap.String("name", name),
				zap.String("type", collectionType),
				zap.String("parentID", parentID))

			remoteOutput, err := remoteCollectionService.Create(ctx, remoteInput)
			if err != nil {
				fmt.Printf("🐞 Error creating remote sub-collection: %v\n", err)
				return
			}

			// Step 2: Download the remote sub-collection to local storage
			remoteID := remoteOutput.Collection.ID.Hex()
			logger.Debug("Downloading remote sub-collection to local storage",
				zap.String("remoteID", remoteID))

			downloadOutput, err := downloadService.Download(ctx, remoteID)
			if err != nil {
				fmt.Printf("🐞 Warning: Created remote sub-collection, but failed to download to local storage: %v\n", err)
				// Continue execution, as the remote collection was successfully created
			}

			// Display success message
			fmt.Println("\n✅ Sub-collection created successfully!")
			fmt.Printf("Collection ID: %s\n", remoteOutput.Collection.ID.Hex())
			fmt.Printf("Collection Type: %s\n", remoteOutput.Collection.Type)
			fmt.Printf("Parent Collection ID: %s\n", parentID)
			fmt.Printf("Created At: %s\n", remoteOutput.Collection.CreatedAt.Format("2006-01-02 15:04:05"))

			if downloadOutput != nil && downloadOutput.Collection != nil {
				fmt.Println("Collection synchronized to local storage.")
			}
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&name, "name", "n", "", "Name of the sub-collection (required)")
	cmd.Flags().StringVarP(&collectionType, "type", "t", "folder", "Type of collection ('folder' or 'album')")
	cmd.Flags().StringVarP(&parentID, "parent", "p", "", "ID of the parent collection (required)")

	// Mark required flags
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("parent")

	return cmd
}
