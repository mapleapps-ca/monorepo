// monorepo/native/desktop/maplefile-cli/cmd/collections/create.go
package collections

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

// createRootCollectionCmd creates a command for creating a root collection
func createRootCollectionCmd(
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
