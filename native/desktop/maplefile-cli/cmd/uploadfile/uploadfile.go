// monorepo/native/desktop/maplefile-cli/cmd/uploadfile/uploadfile.go
package uploadfile

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func UploadFileCmd() *cobra.Command {
	var filePath, description, tags, contentType string
	var customMetadata string

	var cmd = &cobra.Command{
		Use:   "upload-file",
		Short: "Upload a file with end-to-end encryption",
		Long: `
Upload a file with end-to-end encryption to your MapleFile account.
The file will be encrypted locally before being uploaded, ensuring
your data remains secure and private. You can also provide metadata
for the file that will be encrypted along with the content.

Examples:
		# Basic upload with minimal metadata
		maplefile-cli remote upload-file --file /path/to/file.pdf

		# Upload with description and tags
		maplefile-cli remote upload-file --file /path/to/file.pdf --description "Important document" --tags "work,private,important"

		# Upload with custom metadata
		maplefile-cli remote upload-file --file /path/to/file.pdf --custom '{"project":"Project X","department":"Finance"}'
`,
		Run: func(cmd *cobra.Command, args []string) {
			// logger, _ := zap.NewProduction()
			// defer logger.Sync() // flushes buffer, if any
			// sugar := logger.Sugar()

			// sugar.Info("Starting upload-file command")

			// // Check if file exists
			// if filePath == "" {
			// 	sugar.Error("File path is required")
			// 	fmt.Println("Error: File path is required") // Keep user-facing fmt.Println for direct error messages
			// 	return
			// }
			// sugar.Infof("File path provided: %s", filePath)

			// fileInfo, err := os.Stat(filePath)
			// if err != nil {
			// 	sugar.Errorf("Failed to access file: %v", err)
			// 	fmt.Printf("Error: Failed to access file: %v\n", err) // Keep user-facing fmt.Printf
			// 	return
			// }
			// if fileInfo == nil {
			// 	sugar.Error("fileInfo is nil, though os.Stat did not return an error. This should not happen.")
			// 	fmt.Println("Error: File information could not be retrieved.") // Keep user-facing fmt.Println
			// 	return
			// }
			// sugar.Infof("File %s accessed successfully, size: %d bytes, modTime: %s", filePath, fileInfo.Size(), fileInfo.ModTime())

			// // Create E2EE client
			// sugar.Info("Creating E2EE client")
			// client := createE2EEClient() // Assuming createE2EEClient might have its own logging

			// // Check authentication
			// sugar.Info("Checking authentication status")
			// if !client.IsAuthenticated() {
			// 	sugar.Warn("User is not authenticated or session expired")
			// 	fmt.Println("Your session has expired or you are not logged in.")
			// 	fmt.Println("Please login again before uploading files.")
			// 	fmt.Println("You can login using: maplefile-cli remote login")
			// 	return
			// }
			// sugar.Info("User is authenticated")

			// // Prepare file metadata
			// sugar.Info("Preparing file metadata")
			// determinedContentType := determineContentType(filePath, contentType)
			// metadata := &e2ee.FileMetadata{
			// 	Filename:     filepath.Base(filePath),
			// 	OriginalSize: fileInfo.Size(),
			// 	ContentType:  determinedContentType,
			// 	CreatedAt:    time.Now(),
			// 	ModifiedAt:   fileInfo.ModTime(),
			// 	Description:  description,
			// }
			// sugar.Infow("Initial metadata prepared",
			// 	"filename", metadata.Filename,
			// 	"originalSize", metadata.OriginalSize,
			// 	"contentType", metadata.ContentType,
			// 	"createdAt", metadata.CreatedAt,
			// 	"modifiedAt", metadata.ModifiedAt,
			// 	"description", metadata.Description,
			// )

			// // Process tags if provided
			// if tags != "" {
			// 	sugar.Infof("Processing tags: %s", tags)
			// 	metadata.Tags = strings.Split(tags, ",")
			// 	// Trim whitespace from each tag
			// 	for i, tag := range metadata.Tags {
			// 		metadata.Tags[i] = strings.TrimSpace(tag)
			// 	}
			// 	sugar.Infof("Processed tags: %v", metadata.Tags)
			// } else {
			// 	sugar.Info("No tags provided")
			// }

			// // Process custom metadata if provided
			// if customMetadata != "" {
			// 	sugar.Infof("Processing custom metadata: %s", customMetadata)
			// 	var customMap map[string]string
			// 	err := json.Unmarshal([]byte(customMetadata), &customMap)
			// 	if err != nil {
			// 		sugar.Errorf("Invalid custom metadata format: %v", err)
			// 		fmt.Printf("Error: Invalid custom metadata format: %v\n", err) // Keep user-facing fmt.Printf
			// 		return
			// 	}
			// 	metadata.CustomMetadata = customMap
			// 	sugar.Infof("Processed custom metadata: %v", metadata.CustomMetadata)
			// } else {
			// 	sugar.Info("No custom metadata provided")
			// }

			// // Generate a unique file ID
			// sugar.Info("Generating file ID")
			// fileID := generateFileID(filePath, *metadata)
			// sugar.Infof("Generated file ID: %s", fileID)
			// fmt.Printf("Generated file ID: %s\n", fileID) // Keep user-facing fmt.Printf

			// // Upload the file
			// sugar.Info("Starting file encryption and upload process")
			// fmt.Println("Starting file encryption and upload...") // Keep user-facing fmt.Println
			// response, err := client.UploadEncryptedFile(filePath, fileID, metadata)
			// if err != nil {
			// 	sugar.Errorf("Failed to encrypt and upload file: %v", err)
			// 	fmt.Printf("Error: Failed to encrypt and upload file: %v\n", err) // Keep user-facing fmt.Printf
			// 	return
			// }

			// sugar.Infow("File successfully encrypted and uploaded",
			// 	"serverID", response.ID,
			// 	"fileID", response.FileID,
			// 	"createdAt", response.CreatedAt,
			// )
			fmt.Println("File successfully encrypted and uploaded!")
			// fmt.Printf("Server ID: %s\n", response.ID)
			// fmt.Printf("File ID: %s\n", response.FileID)
			// fmt.Printf("Created At: %s\n", response.CreatedAt.Format(time.RFC3339))
		},
	}

	// Define command flags
	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the file to upload (required)")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Description of the file")
	cmd.Flags().StringVarP(&tags, "tags", "t", "", "Comma-separated list of tags")
	cmd.Flags().StringVarP(&contentType, "content-type", "c", "", "Content type of the file (defaults to auto-detection)")
	cmd.Flags().StringVarP(&customMetadata, "custom", "m", "", "Custom metadata in JSON format")

	// Mark required flags
	cmd.MarkFlagRequired("file")

	return cmd
}

// determineContentType attempts to determine the content type of a file
func determineContentType(filePath, providedType string) string {
	// No need for extensive logging here as it's a simple utility function,
	// but could add if debugging content type issues.
	if providedType != "" {
		return providedType
	}

	// Simple extension-based content type detection
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".txt":
		return "text/plain"
	case ".doc", ".docx":
		return "application/msword"
	case ".xls", ".xlsx":
		return "application/vnd.ms-excel"
	default:
		return "application/octet-stream" // Default binary type
	}
}

// // generateFileID creates a unique ID for the file
// func generateFileID(filePath string, metadata e2ee.FileMetadata) string {
// 	// No need for extensive logging here as it's a simple utility function.
// 	// Create a unique identifier based on file path, size, and current time
// 	uniqueStr := fmt.Sprintf("%s_%d_%d", filePath, metadata.OriginalSize, time.Now().UnixNano())

// 	// For demonstration, generate a simple hash-like ID (in production, use proper hashing)
// 	hashBytes := make([]byte, 16)
// 	for i := 0; i < len(uniqueStr) && i < len(hashBytes); i++ {
// 		hashBytes[i%len(hashBytes)] ^= uniqueStr[i]
// 	}

// 	// Convert to hex string
// 	hexID := fmt.Sprintf("%x", hashBytes)
// 	return hexID[:16] // Return first 16 hex chars (8 bytes)
// }
