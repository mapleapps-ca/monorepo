// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/config.go
package config

import (
	"time"

	sbytes "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/security/securebytes"
	sstring "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/security/securestring"
)

type Configuration struct {
	App               AppConfig
	Cache             CacheConf
	DB                DatabaseConfig
	AWS               AWSConfig
	MapleFileMailgun  MailgunConfig
	PaperCloudMailgun MailgunConfig
	Observability     ObservabilityConfig
	Logging           LoggingConfig
}

type CacheConf struct {
	URI string
}

type AppConfig struct {
	DataDirectory            string
	Port                     string
	IP                       string
	AdministrationHMACSecret *sbytes.SecureBytes
	AdministrationSecretKey  *sstring.SecureString
	GeoLiteDBPath            string
	BannedCountries          []string
	BetaAccessCode           string
	Environment              string
	Version                  string
}

type DatabaseConfig struct {
	Hosts             []string
	Keyspace          string
	Consistency       string
	Username          string
	Password          string
	MigrationsPath    string
	ConnectTimeout    time.Duration
	RequestTimeout    time.Duration
	ReplicationFactor int
	MaxRetryAttempts  int
	RetryDelay        time.Duration
}

type MailgunConfig struct {
	APIKey           string
	Domain           string
	APIBase          string
	SenderEmail      string
	MaintenanceEmail string
	FrontendDomain   string
	BackendDomain    string
}

type AWSConfig struct {
	AccessKey  string
	SecretKey  string
	Endpoint   string
	Region     string
	BucketName string
}

// ObservabilityConfig contains configuration for health checks and metrics
type ObservabilityConfig struct {
	Enabled              bool
	Port                 string
	HealthCheckTimeout   time.Duration
	MetricsEnabled       bool
	HealthChecksEnabled  bool
	DetailedHealthChecks bool
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level            string
	Format           string // json or console
	EnableStacktrace bool
	EnableCaller     bool
}

func NewProvider() *Configuration {
	var c Configuration

	//
	// --------- SHARED ------------
	//

	// --- Application section ---
	c.App.DataDirectory = getEnv("BACKEND_APP_DATA_DIRECTORY", true)
	c.App.Port = getEnv("BACKEND_PORT", true)
	c.App.IP = getEnv("BACKEND_IP", false)
	c.App.AdministrationHMACSecret = getSecureBytesEnv("BACKEND_APP_ADMINISTRATION_HMAC_SECRET", false)
	c.App.AdministrationSecretKey = getSecureStringEnv("BACKEND_APP_ADMINISTRATION_SECRET_KEY", false)
	c.App.GeoLiteDBPath = getEnv("BACKEND_APP_GEOLITE_DB_PATH", false)
	c.App.BannedCountries = getStringsArrEnv("BACKEND_APP_BANNED_COUNTRIES", false)
	c.App.BetaAccessCode = getEnv("BACKEND_APP_BETA_ACCESS_CODE", false)
	c.App.Environment = getEnv("ENVIRONMENT", false)
	if c.App.Environment == "" {
		c.App.Environment = "development"
	}
	c.App.Version = getEnv("SERVICE_VERSION", false)
	if c.App.Version == "" {
		c.App.Version = "1.0.0"
	}

	// --- Database section ---
	c.DB = DatabaseConfig{
		Hosts:             getStringsArrEnv("BACKEND_DB_HOSTS", true),
		Keyspace:          getEnv("BACKEND_DB_KEYSPACE", true),
		Consistency:       getEnv("BACKEND_DB_CONSISTENCY", true),
		Username:          getEnv("BACKEND_DB_USERNAME", false),
		Password:          getEnv("BACKEND_DB_PASSWORD", false),
		MigrationsPath:    getEnv("BACKEND_MIGRATIONS_PATH", true),
		ConnectTimeout:    getEnvDuration("BACKEND_DB_CONNECT_TIMEOUT", true),
		RequestTimeout:    getEnvDuration("BACKEND_DB_REQUEST_TIMEOUT", true),
		ReplicationFactor: int(getInt64Env("BACKEND_DB_REPLICATION_FACTOR", true)),
		MaxRetryAttempts:  int(getInt64Env("BACKEND_DB_MAX_RETRY_ATTEMPTS", true)),
		RetryDelay:        getEnvDuration("BACKEND_DB_RETRY_DELAY", true),
	}

	// --- Cache ---
	c.Cache.URI = getEnv("BACKEND_CACHE_URI", true)

	// --- AWS ---
	c.AWS.AccessKey = getEnv("BACKEND_AWS_ACCESS_KEY", true)
	c.AWS.SecretKey = getEnv("BACKEND_AWS_SECRET_KEY", true)
	c.AWS.Endpoint = getEnv("BACKEND_AWS_ENDPOINT", true)
	c.AWS.Region = getEnv("BACKEND_AWS_REGION", true)
	c.AWS.BucketName = getEnv("BACKEND_AWS_BUCKET_NAME", true)

	// --- Observability ---
	c.Observability.Enabled = getEnvBool("BACKEND_OBSERVABILITY_ENABLED", false, true)
	c.Observability.Port = getEnv("BACKEND_OBSERVABILITY_PORT", false)
	if c.Observability.Port == "" {
		c.Observability.Port = "8080"
	}
	c.Observability.HealthCheckTimeout = getEnvDuration("BACKEND_HEALTH_CHECK_TIMEOUT", false)
	if c.Observability.HealthCheckTimeout == 0 {
		c.Observability.HealthCheckTimeout = 30 * time.Second
	}
	c.Observability.MetricsEnabled = getEnvBool("BACKEND_METRICS_ENABLED", false, true)
	c.Observability.HealthChecksEnabled = getEnvBool("BACKEND_HEALTH_CHECKS_ENABLED", false, true)
	c.Observability.DetailedHealthChecks = getEnvBool("BACKEND_DETAILED_HEALTH_CHECKS", false, false)

	// --- Logging ---
	c.Logging.Level = getEnv("LOG_LEVEL", false)
	if c.Logging.Level == "" {
		if c.App.Environment == "production" {
			c.Logging.Level = "info"
		} else {
			c.Logging.Level = "debug"
		}
	}
	c.Logging.Format = getEnv("LOG_FORMAT", false)
	if c.Logging.Format == "" {
		if c.App.Environment == "production" {
			c.Logging.Format = "json"
		} else {
			c.Logging.Format = "console"
		}
	}
	c.Logging.EnableStacktrace = getEnvBool("LOG_ENABLE_STACKTRACE", false, c.App.Environment != "production")
	c.Logging.EnableCaller = getEnvBool("LOG_ENABLE_CALLER", false, true)

	//
	// --------- MapleFile ------------
	//

	// --- Mailgun ---
	c.MapleFileMailgun.APIKey = getEnv("BACKEND_MAPLEFILE_MAILGUN_API_KEY", true)
	c.MapleFileMailgun.Domain = getEnv("BACKEND_MAPLEFILE_MAILGUN_DOMAIN", true)
	c.MapleFileMailgun.APIBase = getEnv("BACKEND_MAPLEFILE_MAILGUN_API_BASE", true)
	c.MapleFileMailgun.SenderEmail = getEnv("BACKEND_MAPLEFILE_MAILGUN_SENDER_EMAIL", true)
	c.MapleFileMailgun.MaintenanceEmail = getEnv("BACKEND_MAPLEFILE_MAILGUN_MAINTENANCE_EMAIL", true)
	c.MapleFileMailgun.FrontendDomain = getEnv("BACKEND_MAPLEFILE_MAILGUN_FRONTEND_DOMAIN", true)
	c.MapleFileMailgun.BackendDomain = getEnv("BACKEND_MAPLEFILE_MAILGUN_BACKEND_DOMAIN", true)

	//
	// --------- PaperCloud ------------
	//

	// --- Mailgun ---
	c.PaperCloudMailgun.APIKey = getEnv("BACKEND_PAPERCLOUD_MAILGUN_API_KEY", true)
	c.PaperCloudMailgun.Domain = getEnv("BACKEND_PAPERCLOUD_MAILGUN_DOMAIN", true)
	c.PaperCloudMailgun.APIBase = getEnv("BACKEND_PAPERCLOUD_MAILGUN_API_BASE", true)
	c.PaperCloudMailgun.SenderEmail = getEnv("BACKEND_PAPERCLOUD_MAILGUN_SENDER_EMAIL", true)
	c.PaperCloudMailgun.MaintenanceEmail = getEnv("BACKEND_PAPERCLOUD_MAILGUN_MAINTENANCE_EMAIL", true)
	c.PaperCloudMailgun.FrontendDomain = getEnv("BACKEND_PAPERCLOUD_MAILGUN_FRONTEND_DOMAIN", true)
	c.PaperCloudMailgun.BackendDomain = getEnv("BACKEND_PAPERCLOUD_MAILGUN_BACKEND_DOMAIN", true)

	return &c
}
