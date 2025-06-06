package blacklist

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createTempFile(t *testing.T, content string) string {
	tmpfile, err := os.CreateTemp("", "blacklist*.json")
	assert.NoError(t, err)

	err = os.WriteFile(tmpfile.Name(), []byte(content), 0644)
	assert.NoError(t, err)

	return tmpfile.Name()
}

func TestReadBlacklistFileContent(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantItems []string
		wantErr   bool
	}{
		{
			name:      "valid json",
			content:   `["192.168.1.1", "10.0.0.1"]`,
			wantItems: []string{"192.168.1.1", "10.0.0.1"},
			wantErr:   false,
		},
		{
			name:      "empty array",
			content:   `[]`,
			wantItems: []string{},
			wantErr:   false,
		},
		{
			name:      "invalid json",
			content:   `invalid json`,
			wantItems: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpfile := createTempFile(t, tt.content)
			defer os.Remove(tmpfile)

			items, err := readBlacklistFileContent(tmpfile)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, items)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantItems, items)
			}
		})
	}

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := readBlacklistFileContent("nonexistent.json")
		assert.Error(t, err)
	})
}

func TestNewProvider(t *testing.T) {
	// Create temporary blacklist files
	ipsContent := `["192.168.1.1", "10.0.0.1"]`
	urlsContent := `["example.com", "malicious.com"]`

	tmpDir, err := os.MkdirTemp("", "blacklist")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	err = os.MkdirAll(filepath.Join(tmpDir, "static/blacklist"), 0755)
	assert.NoError(t, err)

	err = os.WriteFile(filepath.Join(tmpDir, "static/blacklist/ips.json"), []byte(ipsContent), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, "static/blacklist/urls.json"), []byte(urlsContent), 0644)
	assert.NoError(t, err)

	// Change working directory temporarily
	originalWd, err := os.Getwd()
	assert.NoError(t, err)
	err = os.Chdir(tmpDir)
	assert.NoError(t, err)
	defer os.Chdir(originalWd)

	provider := NewProvider()
	assert.NotNil(t, provider)

	// Test IP blacklist
	assert.True(t, provider.IsBannedIPAddress("192.168.1.1"))
	assert.True(t, provider.IsBannedIPAddress("10.0.0.1"))
	assert.False(t, provider.IsBannedIPAddress("172.16.0.1"))

	// Test URL blacklist
	assert.True(t, provider.IsBannedURL("example.com"))
	assert.True(t, provider.IsBannedURL("malicious.com"))
	assert.False(t, provider.IsBannedURL("safe.com"))
}

func TestIsBannedIPAddress(t *testing.T) {
	provider := blacklistProvider{
		bannedIPAddresses: map[string]bool{
			"192.168.1.1": true,
			"10.0.0.1":    true,
		},
	}

	assert.True(t, provider.IsBannedIPAddress("192.168.1.1"))
	assert.True(t, provider.IsBannedIPAddress("10.0.0.1"))
	assert.False(t, provider.IsBannedIPAddress("172.16.0.1"))
}

func TestIsBannedURL(t *testing.T) {
	provider := blacklistProvider{
		bannedURLs: map[string]bool{
			"example.com":   true,
			"malicious.com": true,
		},
	}

	assert.True(t, provider.IsBannedURL("example.com"))
	assert.True(t, provider.IsBannedURL("malicious.com"))
	assert.False(t, provider.IsBannedURL("safe.com"))
}
