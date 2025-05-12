package ipcountryblocker

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
)

// testProvider is a test-specific wrapper that allows access to internal fields
// of the provider struct for verification in tests. This is a common pattern
// when you need to test internal state while keeping the production interface clean.
type testProvider struct {
	Provider           // Embedded interface for normal operations
	internal *provider // Access to internal fields for testing
}

// newTestProvider creates a test provider instance with access to internal fields.
// This allows us to verify the internal state in our tests while maintaining
// encapsulation in production code.
func newTestProvider(cfg *config.Configuration, logger *zap.Logger) testProvider {
	p := NewProvider(cfg, logger)
	return testProvider{
		Provider: p,
		internal: p.(*provider), // Type assertion to get access to internal fields
	}
}

// TestNewProvider verifies that the provider is properly initialized with all
// required components (database connection, blocked countries map, logger).
func TestNewProvider(t *testing.T) {
	// Setup test configuration with path to test database
	cfg := &config.Configuration{
		App: config.AppConfig{
			GeoLiteDBPath:   "../../../static/GeoLite2-Country.mmdb",
			BannedCountries: []string{"US", "CN"},
		},
	}
	// Initialize logger with JSON output for structured test logs
	logger, _ := zap.NewDevelopment()

	// Create test provider and verify internal components
	p := newTestProvider(cfg, logger)
	assert.NotNil(t, p.Provider, "Provider should not be nil")
	assert.NotEmpty(t, p.internal.blockedCountries, "Blocked countries map should be initialized")
	assert.NotNil(t, p.internal.logger, "Logger should be initialized")
	assert.NotNil(t, p.internal.db, "Database connection should be initialized")
	defer p.Close() // Ensure cleanup after test
}

// TestProvider_IsBlockedCountry tests the country blocking functionality with
// various country codes including edge cases like empty and invalid codes.
func TestProvider_IsBlockedCountry(t *testing.T) {
	provider := setupTestProvider(t)
	defer provider.Close()

	// Table-driven test cases covering various scenarios
	tests := []struct {
		name     string
		country  string
		expected bool
	}{
		// Positive test cases - blocked countries
		{
			name:     "blocked country US",
			country:  "US",
			expected: true,
		},
		{
			name:     "blocked country CN",
			country:  "CN",
			expected: true,
		},
		// Negative test cases - allowed countries
		{
			name:     "non-blocked country GB",
			country:  "GB",
			expected: false,
		},
		{
			name:     "non-blocked country JP",
			country:  "JP",
			expected: false,
		},
		// Edge cases
		{
			name:     "empty country code",
			country:  "",
			expected: false,
		},
		{
			name:     "invalid country code",
			country:  "XX",
			expected: false,
		},
		{
			name:     "lowercase country code", // Tests case sensitivity
			country:  "us",
			expected: false,
		},
	}

	// Run each test case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.IsBlockedCountry(tt.country)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestProvider_IsBlockedIP verifies IP blocking functionality using real-world
// IP addresses, including IPv4, IPv6, and various edge cases.
func TestProvider_IsBlockedIP(t *testing.T) {
	provider := setupTestProvider(t)
	defer provider.Close()

	tests := []struct {
		name     string
		ip       net.IP
		expected bool
	}{
		// Known IP addresses from blocked countries
		{
			name:     "blocked IP (US - Google DNS)",
			ip:       net.ParseIP("8.8.8.8"), // Google's primary DNS
			expected: true,
		},
		{
			name:     "blocked IP (US - Google DNS 2)",
			ip:       net.ParseIP("8.8.4.4"), // Google's secondary DNS
			expected: true,
		},
		{
			name:     "blocked IP (CN - Alibaba)",
			ip:       net.ParseIP("223.5.5.5"), // Alibaba DNS
			expected: true,
		},
		// Non-blocked country IPs
		{
			name:     "non-blocked IP (GB)",
			ip:       net.ParseIP("178.62.1.1"),
			expected: false,
		},
		// Edge cases and special scenarios
		{
			name:     "nil IP",
			ip:       nil,
			expected: false,
		},
		{
			name:     "invalid IP format",
			ip:       net.ParseIP("invalid"),
			expected: false,
		},
		{
			name:     "IPv6 address",
			ip:       net.ParseIP("2001:4860:4860::8888"), // Google's IPv6 DNS
			expected: true,
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.IsBlockedIP(ctx, tt.ip)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestProvider_GetCountryCode verifies the country code lookup functionality
// for various IP addresses, including error cases.
func TestProvider_GetCountryCode(t *testing.T) {
	provider := setupTestProvider(t)
	defer provider.Close()

	tests := []struct {
		name        string
		ip          net.IP
		expected    string
		expectError bool
	}{
		// Valid IP addresses with known countries
		{
			name:        "US IP (Google DNS)",
			ip:          net.ParseIP("8.8.8.8"),
			expected:    "US",
			expectError: false,
		},
		// Error cases
		{
			name:        "nil IP",
			ip:          nil,
			expected:    "",
			expectError: true,
		},
		{
			name:        "private IP", // RFC 1918 address
			ip:          net.ParseIP("192.168.1.1"),
			expected:    "",
			expectError: true,
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := provider.GetCountryCode(ctx, tt.ip)
			if tt.expectError {
				assert.Error(t, err, "Should return error for invalid IP")
				assert.Empty(t, code, "Should return empty code on error")
				return
			}
			assert.NoError(t, err, "Should not return error for valid IP")
			assert.Equal(t, tt.expected, code, "Should return correct country code")
		})
	}
}

// TestProvider_Close verifies that the provider properly closes its resources
// and subsequent operations fail as expected.
func TestProvider_Close(t *testing.T) {
	provider := setupTestProvider(t)

	// Verify initial close succeeds
	err := provider.Close()
	assert.NoError(t, err, "Initial close should succeed")

	// Verify operations fail after close
	code, err := provider.GetCountryCode(context.Background(), net.ParseIP("8.8.8.8"))
	assert.Error(t, err, "Operations should fail after close")
	assert.Empty(t, code, "No data should be returned after close")
}

// setupTestProvider is a helper function that creates a properly configured
// provider instance for testing, using the test database path.
func setupTestProvider(t *testing.T) Provider {
	cfg := &config.Configuration{
		App: config.AppConfig{
			GeoLiteDBPath:   "../../../static/GeoLite2-Country.mmdb",
			BannedCountries: []string{"US", "CN"},
		},
	}
	logger, _ := zap.NewDevelopment()
	return NewProvider(cfg, logger)
}
