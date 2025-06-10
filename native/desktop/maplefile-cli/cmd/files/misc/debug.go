// native/desktop/maplefile-cli/cmd/files/misc/debug.go
package misc

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	uc_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collection"
	uc_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
	pkg_crypto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
)

// debugE2EECmd creates a command for debugging E2EE key chain issues
func debugE2EECmd(
	logger *zap.Logger,
	getFileUseCase uc_file.GetFileUseCase,
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase,
	getCollectionUseCase uc_collection.GetCollectionUseCase,
) *cobra.Command {
	var fileID string
	var password string

	var cmd = &cobra.Command{
		Use:   "debug-e2ee",
		Short: "Debug E2EE key chain decryption",
		Long:  `Debug tool to test E2EE key chain decryption step-by-step.`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			if fileID == "" {
				fmt.Println("❌ Error: File ID is required.")
				return
			}

			if password == "" {
				fmt.Println("❌ Error: Password is required.")
				return
			}

			fileObjectID, err := primitive.ObjectIDFromHex(fileID)
			if err != nil {
				fmt.Printf("❌ Error: Invalid file ID format: %v\n", err)
				return
			}

			fmt.Printf("🔍 Debugging E2EE key chain for file: %s\n", fileID)

			// Step 1: Get file metadata
			fmt.Println("\n📁 Step 1: Getting file metadata...")
			file, err := getFileUseCase.Execute(ctx, fileObjectID)
			if err != nil {
				fmt.Printf("❌ Failed to get file: %v\n", err)
				return
			}
			if file == nil {
				fmt.Println("❌ File not found")
				return
			}
			fmt.Printf("✅ File found: %s (Collection: %s)\n", file.Name, file.CollectionID.String())

			// Step 2: Get user
			fmt.Println("\n👤 Step 2: Getting logged in user...")
			user, err := getUserByIsLoggedInUseCase.Execute(ctx)
			if err != nil {
				fmt.Printf("❌ Failed to get user: %v\n", err)
				return
			}
			if user == nil {
				fmt.Println("❌ User not found - please log in first")
				return
			}
			fmt.Printf("✅ User found: %s\n", user.Email)
			fmt.Printf("   Password salt length: %d bytes\n", len(user.PasswordSalt))
			fmt.Printf("   Encrypted master key ciphertext length: %d bytes\n", len(user.EncryptedMasterKey.Ciphertext))
			fmt.Printf("   Encrypted master key nonce length: %d bytes\n", len(user.EncryptedMasterKey.Nonce))

			// Step 3: Get collection
			fmt.Println("\n📂 Step 3: Getting collection...")
			collection, err := getCollectionUseCase.Execute(ctx, file.CollectionID)
			if err != nil {
				fmt.Printf("❌ Failed to get collection: %v\n", err)
				return
			}
			if collection == nil {
				fmt.Println("❌ Collection not found")
				return
			}
			fmt.Printf("✅ Collection found: %s\n", collection.Name)
			if collection.EncryptedCollectionKey != nil {
				fmt.Printf("   Encrypted collection key ciphertext length: %d bytes\n", len(collection.EncryptedCollectionKey.Ciphertext))
				fmt.Printf("   Encrypted collection key nonce length: %d bytes\n", len(collection.EncryptedCollectionKey.Nonce))
			} else {
				fmt.Println("❌ Collection has no encrypted key!")
				return
			}

			// Step 4: Test key derivation
			fmt.Println("\n🔑 Step 4: Testing key derivation...")
			keyEncryptionKey, err := pkg_crypto.DeriveKeyFromPassword(password, user.PasswordSalt)
			if err != nil {
				fmt.Printf("❌ Failed to derive key encryption key: %v\n", err)
				return
			}
			defer pkg_crypto.ClearBytes(keyEncryptionKey)
			fmt.Printf("✅ Key encryption key derived successfully (%d bytes)\n", len(keyEncryptionKey))

			// Step 5: Test master key decryption
			fmt.Println("\n🔓 Step 5: Testing master key decryption...")
			masterKey, err := pkg_crypto.DecryptWithSecretBox(
				user.EncryptedMasterKey.Ciphertext,
				user.EncryptedMasterKey.Nonce,
				keyEncryptionKey,
			)
			if err != nil {
				fmt.Printf("❌ Failed to decrypt master key: %v\n", err)
				fmt.Println("   This usually means the password is incorrect!")
				return
			}
			defer pkg_crypto.ClearBytes(masterKey)
			fmt.Printf("✅ Master key decrypted successfully (%d bytes)\n", len(masterKey))

			// Step 6: Test collection key decryption
			fmt.Println("\n🗝️  Step 6: Testing collection key decryption...")
			collectionKey, err := pkg_crypto.DecryptWithSecretBox(
				collection.EncryptedCollectionKey.Ciphertext,
				collection.EncryptedCollectionKey.Nonce,
				masterKey,
			)
			if err != nil {
				fmt.Printf("❌ Failed to decrypt collection key: %v\n", err)
				return
			}
			defer pkg_crypto.ClearBytes(collectionKey)
			fmt.Printf("✅ Collection key decrypted successfully (%d bytes)\n", len(collectionKey))

			// Step 7: Test file key decryption
			fmt.Println("\n🔐 Step 7: Testing file key decryption...")
			fileKey, err := pkg_crypto.DecryptWithSecretBox(
				file.EncryptedFileKey.Ciphertext,
				file.EncryptedFileKey.Nonce,
				collectionKey,
			)
			if err != nil {
				fmt.Printf("❌ Failed to decrypt file key: %v\n", err)
				return
			}
			defer pkg_crypto.ClearBytes(fileKey)
			fmt.Printf("✅ File key decrypted successfully (%d bytes)\n", len(fileKey))

			fmt.Println("\n🎉 All E2EE decryption steps completed successfully!")
			fmt.Println("   The password and key chain are working correctly.")
			fmt.Println("   If download is still failing, the issue might be with:")
			fmt.Println("   - Network connectivity to the cloud service")
			fmt.Println("   - Presigned URL generation")
			fmt.Println("   - File content decryption format")
		},
	}

	cmd.Flags().StringVarP(&fileID, "file-id", "f", "", "ID of the file to debug (required)")
	cmd.MarkFlagRequired("file-id")
	cmd.Flags().StringVar(&password, "password", "", "Your account password (required)")
	cmd.MarkFlagRequired("password")

	return cmd
}
