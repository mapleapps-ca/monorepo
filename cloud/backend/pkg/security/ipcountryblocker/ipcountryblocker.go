package ipcountryblocker

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/oschwald/geoip2-golang"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
)

// Provider defines the interface for IP-based country blocking operations.
// It provides methods to check if an IP or country is blocked and to retrieve
// country codes for given IP addresses.
type Provider interface {
	// IsBlockedCountry checks if a country is in the blocked list.
	// isoCode must be an ISO 3166-1 alpha-2 country code.
	IsBlockedCountry(isoCode string) bool

	// IsBlockedIP determines if an IP address originates from a blocked country.
	// Returns false for nil IP addresses or if country lookup fails.
	IsBlockedIP(ctx context.Context, ip net.IP) bool

	// GetCountryCode returns the ISO 3166-1 alpha-2 country code for an IP address.
	// Returns an error if the lookup fails or no country is found.
	GetCountryCode(ctx context.Context, ip net.IP) (string, error)

	// Close releases resources associated with the provider.
	Close() error
}

// provider implements the Provider interface using MaxMind's GeoIP2 database.
type provider struct {
	db               *geoip2.Reader
	blockedCountries map[string]struct{} // Uses empty struct to optimize memory
	logger           *zap.Logger
	mu               sync.RWMutex // Protects concurrent access to blockedCountries
}

// NewProvider creates a new IP country blocking provider using the provided configuration.
// It initializes the GeoIP2 database and sets up the blocked countries list.
// Fatally crashes the entire application if the database cannot be opened.
func NewProvider(cfg *config.Configuration, logger *zap.Logger) Provider {
	db, err := geoip2.Open(cfg.App.GeoLiteDBPath)
	if err != nil {
		log.Fatalf("failed to open GeoLite2 DB: %v", err)
	}

	blocked := make(map[string]struct{}, len(cfg.App.BannedCountries))
	for _, country := range cfg.App.BannedCountries {
		blocked[country] = struct{}{}
	}

	logger.Debug("ip blocker initialized",
		zap.String("db_path", cfg.App.GeoLiteDBPath),
		zap.Any("blocked_countries", cfg.App.BannedCountries))

	return &provider{
		db:               db,
		blockedCountries: blocked,
		logger:           logger,
	}
}

// IsBlockedCountry checks if a country code exists in the blocked countries map.
// Thread-safe through RLock.
func (p *provider) IsBlockedCountry(isoCode string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	_, exists := p.blockedCountries[isoCode]
	return exists
}

// IsBlockedIP performs a country lookup for the IP and checks if it's blocked.
// Returns false for nil IPs or failed lookups to fail safely.
func (p *provider) IsBlockedIP(ctx context.Context, ip net.IP) bool {
	if ip == nil {
		return false
	}

	code, err := p.GetCountryCode(ctx, ip)
	if err != nil {
		// Developers Note:
		// Comment this console log as it contributes a `noisy` server log.
		// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		// p.logger.WarnContext(ctx, "failed to get country code",
		// 	zap.Any("ip", ip),
		// 	zap.Any("error", err))
		// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

		// Developers Note:
		// If the country d.n.e. exist that means we will return with `false`
		// indicating this IP address is allowed to access our server. If this
		// is concerning then you might set this to `true` to block on all
		// IP address which are not categorized by country.
		return false
	}

	return p.IsBlockedCountry(code)
}

// GetCountryCode performs a GeoIP2 database lookup to determine an IP's country.
// Returns an error if the lookup fails or no country is found.
func (p *provider) GetCountryCode(ctx context.Context, ip net.IP) (string, error) {
	record, err := p.db.Country(ip)
	if err != nil {
		return "", fmt.Errorf("lookup country: %w", err)
	}

	if record == nil || record.Country.IsoCode == "" {
		return "", fmt.Errorf("no country found for IP: %v", ip)
	}

	return record.Country.IsoCode, nil
}

// Close cleanly shuts down the GeoIP2 database connection.
func (p *provider) Close() error {
	return p.db.Close()
}
