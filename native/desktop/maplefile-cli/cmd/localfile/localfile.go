// Package localfile provides commands for managing local files
package localfile

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/localfile"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
)

// LocalFileCmd creates a command for local file operations
func LocalFileCmd(
	importService localfile.ImportService,
	logger *zap.Logger,
) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "localfile",
		Short: "Manage local files",
		Long:  `Import and manage files on the local filesystem without synchronization.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Show help when no subcommand is specified
			cmd.Help()
		},
	}

	// Add file management subcommands
	cmd.AddCommand(createLocalFileCmd(importService, logger))

	return cmd
}

// generateRandomHexID generates a random hex string of the specified length
func generateRandomHexID(length int) (string, error) {
	// For simplicity, we'll use half the length since each byte becomes 2 hex chars
	randomBytes, err := crypto.GenerateRandomBytes(length / 2)
	if err != nil {
		return "", err
	}

	// Convert to hex, this will be twice the length of the input bytes
	hexID := fmt.Sprintf("%x", randomBytes)

	// Ensure we have exactly the requested length
	if len(hexID) > length {
		hexID = hexID[:length]
	}

	return hexID, nil
}

// base64Encode simulates encryption by encoding to base64
func base64Encode(input string) string {
	return crypto.ToBase64([]byte(input))
}

// getMimeType attempts to determine MIME type from filename
func getMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	// This is a very simplified MIME type detection
	// In a real application, you'd use a more comprehensive method
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".txt":
		return "text/plain"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".pptx":
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	default:
		return "application/octet-stream"
	}
}
