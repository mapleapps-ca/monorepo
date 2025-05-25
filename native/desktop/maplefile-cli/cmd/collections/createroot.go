// monorepo/native/desktop/maplefile-cli/cmd/collections/create.go
package collections

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/collection"
)

// createRootCollectionCmd creates a command for creating a root collection
func createRootCollectionCmd(
	createCollectionService collection.CreateService,
	logger *zap.Logger,
) *cobra.Command {
	var name, collectionType string
	var password string

	var cmd = &cobra.Command{
		Use:   "create",
		Short: "Create a new root collection",
		Long: `
Create a new root collection to store files.

A collection is an encrypted container for your files. Root collections appear
at the top level of your collection hierarchy.

Examples:
  # Create a folder collection
  maplefile-cli collections create --name "My Documents" --type folder --password 1234567890

  # Create an album collection
  maplefile-cli collections create --name "Photo Album" --type album --password 1234567890
`,
		Run: func(cmd *cobra.Command, args []string) {
			// Validate required fields
			if name == "" {
				fmt.Println("🐞 Error: Collection name is required.")
				fmt.Println("Use --name flag to specify the name of your collection.")
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
				OwnerID:        primitive.NewObjectID(), // Placeholder - service will use authenticated user ID
				// ParentID is empty for root collections
			}

			// Call the create service
			output, err := createCollectionService.Create(cmd.Context(), input, password)
			if err != nil {
				fmt.Printf("🐞 Error creating collection: %v\n", err)
				if strings.Contains(err.Error(), "incorrect password") {
					fmt.Printf("❌ Error: Incorrect password. Please check your password and try again.\n")
				} else {
					logger.Error("Failed to create collection",
						zap.String("name", name),
						zap.String("type", collectionType),
						zap.Error(err))
				}
				return
			}

			// Display success message
			fmt.Printf("✅ Successfully created collection!\n\n")
			fmt.Printf("Collection Details:\n")

			// Handle collection name display
			displayName := output.Collection.Name
			if displayName == "" {
				displayName = "[Encrypted]"
			}

			fmt.Printf("  Name: %s\n", displayName)
			fmt.Printf("  Type: %s\n", output.Collection.CollectionType)
			fmt.Printf("  ID: %s\n", output.Collection.ID.Hex())
			fmt.Printf("  Created: %s\n", output.Collection.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("  Sync Status: %s\n", output.Collection.SyncStatus.String())

			logger.Info("Collection created successfully",
				zap.String("name", name),
				zap.String("id", output.Collection.ID.Hex()),
				zap.String("type", collectionType))
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&name, "name", "n", "", "Name of the collection (required)")
	cmd.Flags().StringVarP(&collectionType, "type", "t", "folder", "Type of collection ('folder' or 'album')")

	// Mark required flags
	cmd.MarkFlagRequired("name")
	cmd.Flags().StringVar(&password, "password", "", "Your account password (required for E2EE)")
	cmd.MarkFlagRequired("password")

	return cmd
}
