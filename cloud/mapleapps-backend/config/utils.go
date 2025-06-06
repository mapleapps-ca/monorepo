package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gocql/gocql"
	sbytes "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/security/securebytes"
	sstring "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/security/securestring"
)

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

func getEnvDuration(key string, required bool) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		if required {
			log.Fatalf("Environment variable not found: %s", key)
		}
		// If not required and value is empty, return zero duration (0)
		return 0
	}

	// If value is not empty, try to parse it
	duration, err := time.ParseDuration(value)
	if err != nil {
		// Log error about invalid value for the key
		log.Fatalf("Invalid time.Duration value '%s' for environment variable %s: %v", value, key, err)
	}

	// If parsing succeeds, return the parsed duration
	return duration
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
