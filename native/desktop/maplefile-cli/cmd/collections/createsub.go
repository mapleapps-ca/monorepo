// monorepo/native/desktop/maplefile-cli/cmd/collections/create.go
package collections

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/collection"
)

// createSubCollectionCmd creates a command for creating a sub-collection
func createSubCollectionCmd(
	createCollectionService collection.CreateService,
	logger *zap.Logger,
) *cobra.Command {
	var name, collectionType, parentID string
	var password string

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
			pid, err := primitive.ObjectIDFromHex(parentID)
			if err != nil {
				fmt.Printf("🐞 Error: Invalid parent ID format: %v\n", err)
				return
			}

			if password == "" {
				fmt.Println("❌ Error: Password is required for E2EE operations.")
				fmt.Println("Use --password flag to specify your account password.")
				return
			}

			// Create the service input
			// Note: OwnerID is required by the service interface but will be overridden
			// with the authenticated user ID from the service's internal logic
			input := &collection.CreateInput{
				Name:           name,
				CollectionType: collectionType,
				ParentID:       pid,
				OwnerID:        primitive.NewObjectID(), // Placeholder - service will use authenticated user ID
			}

			// Call the create service
			output, err := createCollectionService.Create(cmd.Context(), input, password)
			if err != nil {
				fmt.Printf("🐞 Error creating sub-collection: %v\n", err)
				logger.Error("Failed to create sub-collection",
					zap.String("name", name),
					zap.String("type", collectionType),
					zap.String("parentID", parentID),
					zap.Error(err))
				return
			}

			// Display success message
			fmt.Printf("✅ Successfully created sub-collection!\n\n")
			fmt.Printf("Collection Details:\n")

			// Handle collection name display
			displayName := output.Collection.Name
			if displayName == "" {
				displayName = "[Encrypted]"
			}

			fmt.Printf("  Name: %s\n", displayName)
			fmt.Printf("  Type: %s\n", output.Collection.CollectionType)
			fmt.Printf("  ID: %s\n", output.Collection.ID.Hex())
			fmt.Printf("  Parent ID: %s\n", output.Collection.ParentID.Hex())
			fmt.Printf("  Created: %s\n", output.Collection.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("  Sync Status: %v\n", output.Collection.SyncStatus)

			logger.Info("Sub-collection created successfully",
				zap.String("name", name),
				zap.String("id", output.Collection.ID.Hex()),
				zap.String("type", collectionType),
				zap.String("parentID", parentID))
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&name, "name", "n", "", "Name of the sub-collection (required)")
	cmd.Flags().StringVarP(&collectionType, "type", "t", "folder", "Type of collection ('folder' or 'album')")
	cmd.Flags().StringVarP(&parentID, "parent", "p", "", "ID of the parent collection (required)")

	// Mark required flags
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("parent")
	cmd.Flags().StringVar(&password, "password", "", "Your account password (required for E2EE)")
	cmd.MarkFlagRequired("password")

	return cmd
}
