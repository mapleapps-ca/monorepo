// cmd/collections/create.go - Clean unified create command
package collections

import (
	"fmt"
	"strings"

	"github.com/gocql/gocql"
	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/collection"
)

// createCmd creates a unified command for creating both root and sub-collections
func createCmd(
	createCollectionService collection.CreateService,
	logger *zap.Logger,
) *cobra.Command {
	var collectionType, parentID string
	var password string

	var cmd = &cobra.Command{
		Use:   "create NAME",
		Short: "Create a new collection",
		Long: `
Create a new collection to store files.

By default, creates a root-level collection. Use --parent to create a sub-collection
within an existing collection.

Collection types:
  • folder: General file storage (default)
  • album:  Photo/media collections

Examples:
  # Create a root collection
  maplefile-cli collections create "My Documents" --password 1234567890

  # Create a sub-collection
  maplefile-cli collections create "Project Files" --parent 507f1f77bcf86cd799439011 --password 1234567890

  # Create an album collection
  maplefile-cli collections create "Vacation Photos" --type album --password 1234567890
`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]

			if password == "" {
				fmt.Println("❌ Error: Password is required for E2EE operations.")
				fmt.Println("Use --password flag to specify your account password.")
				return
			}

			// Set default collection type
			if collectionType == "" {
				collectionType = "folder"
			}

			// Validate collection type
			if collectionType != "folder" && collectionType != "album" {
				fmt.Printf("🐞 Error: Invalid collection type: %s\n", collectionType)
				fmt.Println("Collection type must be either 'folder' or 'album'.")
				return
			}

			// Handle parent ID if provided
			var parentObjectID gocql.UUID
			var isSubCollection bool

			if parentID != "" {
				var err error
				parentObjectID, err = primitive.ObjectIDFromHex(parentID)
				if err != nil {
					fmt.Printf("🐞 Error: Invalid parent ID format: %v\n", err)
					return
				}
				isSubCollection = true
			}

			// Create service input
			input := &collection.CreateInput{
				Name:           name,
				CollectionType: collectionType,
				OwnerID:        primitive.NewObjectID(), // Service will use authenticated user ID
			}

			if isSubCollection {
				input.ParentID = parentObjectID
				fmt.Printf("📁 Creating sub-collection '%s'...\n", name)
			} else {
				fmt.Printf("📁 Creating root collection '%s'...\n", name)
			}

			// Execute create
			output, err := createCollectionService.Create(cmd.Context(), input, password)
			if err != nil {
				fmt.Printf("🐞 Error creating collection: %v\n", err)
				if strings.Contains(err.Error(), "incorrect password") {
					fmt.Printf("❌ Error: Incorrect password. Please check your password and try again.\n")
				} else if strings.Contains(err.Error(), "parent not found") {
					fmt.Printf("❌ Error: Parent collection not found. Please check the parent ID.\n")
				} else if strings.Contains(err.Error(), "permission") {
					fmt.Printf("❌ Error: You don't have permission to create collections in the specified parent.\n")
				}
				return
			}

			// Success message
			if isSubCollection {
				fmt.Printf("✅ Successfully created sub-collection!\n\n")
			} else {
				fmt.Printf("✅ Successfully created root collection!\n\n")
			}

			// Display collection details
			displayName := output.Collection.Name
			if displayName == "" {
				displayName = "[Encrypted]"
			}

			fmt.Printf("📋 Collection Details:\n")
			fmt.Printf("  📁 Name: %s\n", displayName)
			fmt.Printf("  🏷️  Type: %s\n", output.Collection.CollectionType)
			fmt.Printf("  🆔 ID: %s\n", output.Collection.ID.String())

			if isSubCollection {
				fmt.Printf("  📂 Parent ID: %s\n", output.Collection.ParentID.String())
			}

			fmt.Printf("  📅 Created: %s\n", output.Collection.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("  🔄 Sync Status: %s\n", output.Collection.SyncStatus.String())

			// Next steps
			fmt.Printf("\n💡 Next steps:\n")
			fmt.Printf("   • Add files: maplefile-cli files add PATH --collection %s --password PASSWORD\n", output.Collection.ID.String())
			fmt.Printf("   • List collections: maplefile-cli collections list\n")
			if !isSubCollection {
				fmt.Printf("   • Create sub-collection: maplefile-cli collections create 'Sub Name' --parent %s --password PASSWORD\n", output.Collection.ID.String())
			}

			logger.Info("Collection created successfully",
				zap.String("name", name),
				zap.String("id", output.Collection.ID.String()),
				zap.String("type", collectionType),
				zap.Bool("isSubCollection", isSubCollection))
		},
	}

	// Flags
	cmd.Flags().StringVarP(&collectionType, "type", "t", "folder", "Type of collection ('folder' or 'album')")
	cmd.Flags().StringVarP(&parentID, "parent", "p", "", "Parent collection ID (creates sub-collection)")
	cmd.Flags().StringVar(&password, "password", "", "Your account password (required for E2EE)")
	cmd.MarkFlagRequired("password")

	return cmd
}
