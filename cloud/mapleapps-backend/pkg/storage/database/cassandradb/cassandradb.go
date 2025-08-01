package cassandradb

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/gocql/gocql"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	c "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
)

// CassandraDB wraps the gocql session with additional functionality
// Enhanced for better compatibility with Cassandra 4.1+ and future 5.x support
type CassandraDB struct {
	Session *gocql.Session
	config  c.DatabaseConfig
}

// NewCassandraConnection establishes a connection to Cassandra cluster with enhanced configuration
// This version includes specific optimizations for Cassandra 4.1 and preparation for 5.x
func NewCassandraConnection(cfg *config.Configuration) (*gocql.Session, error) {
	dbConfig := cfg.DB

	var session *gocql.Session
	var err error

	log.Printf("Starting Cassandra connection process...")
	log.Printf("Target (raw) hosts: %v", dbConfig.Hosts)
	log.Printf("Target keyspace: %s", dbConfig.Keyspace)

	// Resolve hostnames to IP addresses to support Docker container names.
	resolvedHosts := make([]string, 0, len(dbConfig.Hosts))
	for _, host := range dbConfig.Hosts {
		// The following block of code will be used to resolve the dns of our
		// other docker container to get the node's ip address.
		ips, err := net.LookupIP(host)
		if err != nil {
			log.Printf("Warning: failed to lookup IP for host '%s': %v. This host will be skipped.", host, err)
			continue
		}
		if len(ips) > 0 {
			// Use the first resolved IP address.
			ipStr := ips[0].String()
			resolvedHosts = append(resolvedHosts, ipStr)
			log.Printf("Resolved host '%s' to '%s'", host, ipStr)
		}
	}

	if len(resolvedHosts) == 0 {
		return nil, fmt.Errorf("failed to resolve any Cassandra hosts from the provided list: %v", dbConfig.Hosts)
	}
	log.Printf("Target (resolved) hosts: %v", resolvedHosts)

	// Attempt connection with enhanced retry logic
	for attempt := 1; attempt <= dbConfig.MaxRetryAttempts; attempt++ {
		log.Printf("Cassandra connection attempt %d/%d", attempt, dbConfig.MaxRetryAttempts)

		// Create cluster configuration with enhanced settings
		cluster := gocql.NewCluster(resolvedHosts...)

		// Configure authentication if credentials are provided
		if dbConfig.Username != "" && dbConfig.Password != "" {
			cluster.Authenticator = gocql.PasswordAuthenticator{
				Username: dbConfig.Username,
				Password: dbConfig.Password,
			}
			log.Printf("Using authentication for user: %s", dbConfig.Username)
		}

		// Configure timeouts using the actual gocql API
		// ConnectTimeout controls how long to wait when establishing connections
		cluster.ConnectTimeout = dbConfig.ConnectTimeout
		// Timeout controls the default timeout for queries and other operations
		cluster.Timeout = dbConfig.RequestTimeout

		// Configure consistency level with validation
		consistency, parseErr := parseConsistency(dbConfig.Consistency)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid consistency level '%s': %w", dbConfig.Consistency, parseErr)
		}
		cluster.Consistency = consistency
		log.Printf("Using consistency level: %s", dbConfig.Consistency)

		// Protocol version configuration - crucial for version compatibility
		// Start with v4 (widely supported) and let negotiation handle upgrades
		cluster.ProtoVersion = 4
		log.Printf("Starting with protocol version: %d", cluster.ProtoVersion)

		// Enhanced retry policy for better reliability
		// This helps handle temporary network issues and overloaded nodes
		cluster.RetryPolicy = &gocql.ExponentialBackoffRetryPolicy{
			Min:        time.Millisecond * 100,
			Max:        time.Second * 10,
			NumRetries: 3,
		}

		// Enable compression for better network efficiency
		// Snappy compression reduces network traffic without much CPU overhead
		cluster.Compressor = &gocql.SnappyCompressor{}

		// Discovery configuration - these help the driver find and track all nodes
		cluster.DisableInitialHostLookup = false // Allow discovery of other nodes
		cluster.IgnorePeerAddr = false           // Use peer addresses for connections

		// Enable compression for better network efficiency
		cluster.Compressor = &gocql.SnappyCompressor{}

		// Discovery configuration
		cluster.DisableInitialHostLookup = false // Allow discovery
		cluster.IgnorePeerAddr = false           // Use peer addresses

		// First, connect without specifying a keyspace to create it if needed
		log.Printf("Attempting initial connection to create keyspace if needed...")
		session, err = cluster.CreateSession()
		if err == nil {
			log.Printf("Initial connection successful, checking/creating keyspace...")

			// Create keyspace with enhanced replication settings
			if createErr := createKeyspaceIfNotExists(session, dbConfig.Keyspace, dbConfig.ReplicationFactor); createErr != nil {
				session.Close()
				return nil, fmt.Errorf("failed to create keyspace '%s': %w", dbConfig.Keyspace, createErr)
			}

			// Close and reconnect with the keyspace specified
			session.Close()
			log.Printf("Reconnecting with keyspace: %s", dbConfig.Keyspace)
			cluster.Keyspace = dbConfig.Keyspace
			session, err = cluster.CreateSession()
			if err == nil {
				// Enhanced verification - ping the database and check schema
				if pingErr := pingDatabaseWithValidation(session, dbConfig.Keyspace); pingErr != nil {
					session.Close()
					err = fmt.Errorf("database validation failed: %w", pingErr)
				} else {
					log.Printf("Successfully connected to Cassandra cluster!")
					// log.Printf("Protocol version negotiated: %d", session.Protocol())
					log.Printf("Connected to keyspace: %s", dbConfig.Keyspace)
					break
				}
			}
		}

		// Connection failed, log detailed error and wait before retrying
		if attempt < dbConfig.MaxRetryAttempts {
			log.Printf("Connection attempt %d failed: %v", attempt, err)
			log.Printf("Retrying in %v...", dbConfig.RetryDelay)
			time.Sleep(dbConfig.RetryDelay)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to Cassandra after %d attempts: %w", dbConfig.MaxRetryAttempts, err)
	}

	return session, nil
}

// createKeyspaceIfNotExists creates the application keyspace with enhanced replication strategy
func createKeyspaceIfNotExists(session *gocql.Session, keyspace string, replicationFactor int) error {
	log.Printf("Checking if keyspace '%s' exists...", keyspace)

	// Check if keyspace already exists
	var keyspaceName string
	checkQuery := `SELECT keyspace_name FROM system_schema.keyspaces WHERE keyspace_name = ?`
	err := session.Query(checkQuery, keyspace).Scan(&keyspaceName)

	if err == gocql.ErrNotFound {
		log.Printf("Keyspace '%s' does not exist, creating it...", keyspace)

		// Build the CREATE KEYSPACE query with enhanced replication strategy
		// Using SimpleStrategy for development, but prepared for NetworkTopologyStrategy
		createQuery := fmt.Sprintf(`
			CREATE KEYSPACE IF NOT EXISTS %s
			WITH REPLICATION = {
				'class': 'SimpleStrategy',
				'replication_factor': %d
			}
			AND DURABLE_WRITES = true`, keyspace, replicationFactor)

		if execErr := session.Query(createQuery).Exec(); execErr != nil {
			return fmt.Errorf("failed to create keyspace %s with replication factor %d: %w", keyspace, replicationFactor, execErr)
		}

		log.Printf("Successfully created keyspace '%s' with replication factor %d", keyspace, replicationFactor)
	} else if err != nil {
		return fmt.Errorf("failed to check keyspace existence: %w", err)
	} else {
		log.Printf("Keyspace '%s' already exists", keyspace)
	}

	return nil
}

// pingDatabaseWithValidation performs enhanced connectivity and schema validation
func pingDatabaseWithValidation(session *gocql.Session, keyspace string) error {
	// Test 1: Basic connectivity
	var result string
	if err := session.Query("SELECT uuid() FROM system.local").Scan(&result); err != nil {
		return fmt.Errorf("basic connectivity test failed: %w", err)
	}

	if result == "" {
		return fmt.Errorf("basic connectivity test returned empty result")
	}

	log.Printf("Basic connectivity test passed (UUID: %s)", result[:8]+"...") // Show partial UUID

	// Test 2: Keyspace accessibility
	var keyspaceName string
	keyspaceQuery := `SELECT keyspace_name FROM system_schema.keyspaces WHERE keyspace_name = ?`
	if err := session.Query(keyspaceQuery, keyspace).Scan(&keyspaceName); err != nil {
		return fmt.Errorf("keyspace accessibility test failed: %w", err)
	}

	log.Printf("Keyspace accessibility test passed for: %s", keyspaceName)

	// Test 3: Protocol version compatibility
	var releaseVersion string
	if err := session.Query("SELECT release_version FROM system.local").Scan(&releaseVersion); err != nil {
		return fmt.Errorf("version compatibility test failed: %w", err)
	}

	log.Printf("Connected to Cassandra version: %s", releaseVersion)
	// log.Printf("Using protocol version: %d", session.Protocol())

	return nil
}

// Close terminates the database connection
func (db *CassandraDB) Close() {
	if db.Session != nil {
		db.Session.Close()
	}
}

// Health checks if the database connection is still alive with enhanced validation
func (db *CassandraDB) Health() error {
	// Quick health check using a simple query
	var timestamp time.Time
	err := db.Session.Query("SELECT now() FROM system.local").Scan(&timestamp)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	// Validate that we got a reasonable timestamp (within last minute)
	now := time.Now()
	if timestamp.Before(now.Add(-time.Minute)) || timestamp.After(now.Add(time.Minute)) {
		return fmt.Errorf("health check returned suspicious timestamp: %v (current: %v)", timestamp, now)
	}

	return nil
}

// parseConsistency converts string consistency level to gocql constant with enhanced validation
func parseConsistency(consistency string) (gocql.Consistency, error) {
	// Normalize input (handle case variations and common aliases)
	switch consistency {
	case "any", "ANY":
		return gocql.Any, nil
	case "one", "ONE":
		return gocql.One, nil
	case "two", "TWO":
		return gocql.Two, nil
	case "three", "THREE":
		return gocql.Three, nil
	case "quorum", "QUORUM":
		return gocql.Quorum, nil
	case "all", "ALL":
		return gocql.All, nil
	case "local_quorum", "LOCAL_QUORUM", "localquorum":
		return gocql.LocalQuorum, nil
	case "each_quorum", "EACH_QUORUM", "eachquorum":
		return gocql.EachQuorum, nil
	case "local_one", "LOCAL_ONE", "localone":
		return gocql.LocalOne, nil
	// case "local_serial", "LOCAL_SERIAL", "localserial":
	// 	return gocql.LocalSerial, nil
	// case "serial", "SERIAL":
	// 	return gocql.Serial, nil
	default:
		// Provide helpful error message with valid options
		validOptions := []string{
			"any", "one", "two", "three", "quorum", "all",
			"local_quorum", "each_quorum", "local_one", "local_serial", "serial",
		}
		return gocql.Quorum, fmt.Errorf("unknown consistency level '%s'. Valid options: %v", consistency, validOptions)
	}
}
