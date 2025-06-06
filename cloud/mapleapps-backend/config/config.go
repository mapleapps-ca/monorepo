package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gocql/gocql"

	sbytes "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/security/securebytes"
	sstring "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/security/securestring"
)

type Configuration struct {
	App               AppConfig
	Cache             CacheConf
	DB                CassandraDBConfig
	AWS               AWSConfig
	MapleFileMailgun  MailgunConfig
	PaperCloudMailgun MailgunConfig
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
}

type CassandraDBConfig struct {
	Hosts                     []string
	KeyspaceReplicationFactor int64
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

	// --- Database section ---
	c.DB.Hosts = getStringsArrEnv("BACKEND_DB_HOSTS", true)
	c.DB.KeyspaceReplicationFactor = getInt64Env("BACKEND_DB_KEYSPACE_REPLICATION_FACTOR", true)

	// --- Cache ---
	c.Cache.URI = getEnv("BACKEND_CACHE_URI", true)

	// --- AWS ---
	c.AWS.AccessKey = getEnv("BACKEND_AWS_ACCESS_KEY", true)
	c.AWS.SecretKey = getEnv("BACKEND_AWS_SECRET_KEY", true)
	c.AWS.Endpoint = getEnv("BACKEND_AWS_ENDPOINT", true)
	c.AWS.Region = getEnv("BACKEND_AWS_REGION", true)
	c.AWS.BucketName = getEnv("BACKEND_AWS_BUCKET_NAME", true)

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

func getEnv(key string, required bool) string {
	value := os.Getenv(key)
	if required && value == "" {
		log.Fatalf("Environment variable not found: %s", key)
	}
	return value
}

func getSecureStringEnv(key string, required bool) *sstring.SecureString {
	value := os.Getenv(key)
	if required && value == "" {
		log.Fatalf("Environment variable not found: %s", key)
	}
	ss, err := sstring.NewSecureString(value)
	if ss == nil && required == false {
		return nil
	}
	if err != nil {
		log.Fatalf("Environment variable failed to secure: %v", err)
	}
	return ss
}

func getBytesEnv(key string, required bool) []byte {
	value := os.Getenv(key)
	if required && value == "" {
		log.Fatalf("Environment variable not found: %s", key)
	}
	return []byte(value)
}

func getSecureBytesEnv(key string, required bool) *sbytes.SecureBytes {
	value := getBytesEnv(key, required)
	sb, err := sbytes.NewSecureBytes(value)
	if sb == nil && required == false {
		return nil
	}
	if err != nil {
		log.Fatalf("Environment variable failed to secure: %v", err)
	}
	return sb
}

func getEnvBool(key string, required bool, defaultValue bool) bool {
	valueStr := getEnv(key, required)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		log.Fatalf("Invalid boolean value for environment variable %s", key)
	}
	return value
}

func getStringsArrEnv(key string, required bool) []string {
	value := os.Getenv(key)
	if required && value == "" {
		log.Fatalf("Environment variable not found: %s", key)
	}
	return strings.Split(value, ",")
}

func getUint64Env(key string, required bool) uint64 {
	value := os.Getenv(key)
	if required && value == "" {
		log.Fatalf("Environment variable not found: %s", key)
	}
	valueUint64, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		log.Fatalf("Invalid uint64 value for environment variable %s", key)
	}
	return valueUint64
}

func getInt64Env(key string, required bool) int64 {
	value := os.Getenv(key)
	if required && value == "" {
		log.Fatalf("Environment variable not found: %s", key)
	}
	valueInt64, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		log.Fatalf("Invalid int64 value for environment variable %s", key)
	}
	return valueInt64
}

func getEnvCassandraUUID(key string, required bool) gocql.UUID {
	value := os.Getenv(key)
	if required && value == "" {
		log.Fatalf("Environment variable not found: %s", key)
	}
	objectID, err := gocql.ParseUUID(value)
	if err != nil {
		log.Fatalf("Invalid cassandra uuid value for environment variable %s", key)
	}
	return objectID
}
