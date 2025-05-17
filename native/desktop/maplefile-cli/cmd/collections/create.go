// monorepo/native/desktop/maplefile-cli/cmd/collections/healthcheck.go
package collections

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	collectionService "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/collection"
)

// createRootCollectionCmd creates a command for creating a root collection
func createRootCollectionCmd(collectionSvc collectionService.CollectionService, logger *zap.Logger) *cobra.Command {
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

			input := collectionService.CreateCollectionInput{
				Name:           name,
				CollectionType: collectionType,
			}

			// Create the collection
			output, err := collectionSvc.CreateRootCollection(ctx, input)
			if err != nil {
				fmt.Printf("🐞 Error creating collection: %v\n", err)
				return
			}

			// Display success message
			fmt.Println("\n✅ Collection created successfully!")
			fmt.Printf("Collection ID: %s\n", output.Collection.ID.Hex())
			fmt.Printf("Collection Type: %s\n", output.Collection.Type)
			fmt.Printf("Created At: %s\n", output.Collection.CreatedAt.Format("2006-01-02 15:04:05"))
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&name, "name", "n", "", "Name of the collection (required)")
	cmd.Flags().StringVarP(&collectionType, "type", "t", "folder", "Type of collection ('folder' or 'album')")

	// Mark required flags
	cmd.MarkFlagRequired("name")

	return cmd
}

// cmd/collections/collections.go
// Add to the existing file

// createSubCollectionCmd creates a command for creating a sub-collection
func createSubCollectionCmd(collectionSvc collectionService.CollectionService, logger *zap.Logger) *cobra.Command {
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

			input := collectionService.CreateSubCollectionInput{
				Name:           name,
				CollectionType: collectionType,
				ParentID:       parentID,
			}

			// Create the sub-collection
			output, err := collectionSvc.CreateSubCollection(ctx, input)
			if err != nil {
				fmt.Printf("🐞 Error creating sub-collection: %v\n", err)
				return
			}

			// Display success message
			fmt.Println("\n✅ Sub-collection created successfully!")
			fmt.Printf("Collection ID: %s\n", output.Collection.ID.Hex())
			fmt.Printf("Collection Type: %s\n", output.Collection.Type)
			fmt.Printf("Parent Collection ID: %s\n", output.Collection.ParentID.Hex())
			fmt.Printf("Created At: %s\n", output.Collection.CreatedAt.Format("2006-01-02 15:04:05"))
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
