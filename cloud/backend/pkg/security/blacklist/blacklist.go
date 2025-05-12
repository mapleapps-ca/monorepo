package blacklist

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// Provider provides an interface for abstracting time.
type Provider interface {
	IsBannedIPAddress(ipAddress string) bool
	IsBannedURL(url string) bool
}

type blacklistProvider struct {
	bannedIPAddresses map[string]bool
	bannedURLs        map[string]bool
}

// readBlacklistFileContent reads the contents of the blacklist file and returns
// the list of banned items (ex: IP, URLs, etc).
func readBlacklistFileContent(filePath string) ([]string, error) {
	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file %s does not exist", filePath)
	}

	// Read the file contents
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", filePath, err)
	}

	// Parse the JSON content as a list of IPs
	var ips []string
	if err := json.Unmarshal(data, &ips); err != nil {
		return nil, fmt.Errorf("failed to parse JSON file %s: %v", filePath, err)
	}

	return ips, nil
}

// NewProvider Provider contructor that returns the default time provider.
func NewProvider() Provider {
	bannedIPAddresses := make(map[string]bool)
	bannedIPAddressesFilePath := "static/blacklist/ips.json"
	ips, err := readBlacklistFileContent(bannedIPAddressesFilePath)
	if err == nil { // Aka: if the file exists...
		for _, ip := range ips {
			bannedIPAddresses[ip] = true
		}
	}

	bannedURLs := make(map[string]bool)
	bannedURLsFilePath := "static/blacklist/urls.json"
	urls, err := readBlacklistFileContent(bannedURLsFilePath)
	if err == nil { // Aka: if the file exists...
		for _, url := range urls {
			bannedURLs[url] = true
		}
	}

	return blacklistProvider{
		bannedIPAddresses: bannedIPAddresses,
		bannedURLs:        bannedURLs,
	}
}

func (p blacklistProvider) IsBannedIPAddress(ipAddress string) bool {
	return p.bannedIPAddresses[ipAddress]
}

func (p blacklistProvider) IsBannedURL(url string) bool {
	return p.bannedURLs[url]
}
