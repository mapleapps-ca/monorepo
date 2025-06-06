package cassandradb

import (
	"fmt"
	"log"
	"time"

	"github.com/gocql/gocql"

	c "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
)

// CassandraDB wraps the gocql session with additional functionality
// This abstraction makes it easier to mock for testing and add common functionality
type CassandraDB struct {
	Session *gocql.Session
	config  c.DatabaseConfig
}

// NewCassandraConnection establishes a connection to Cassandra cluster with advanced configuration
// This function incorporates production-ready patterns like retry logic, token-aware routing, and dynamic keyspace creation
func NewCassandraConnection(cfg c.DatabaseConfig) (*CassandraDB, error) {
	var session *gocql.Session
	var err error

	// Attempt connection with retry logic - essential for production deployments
	// Cassandra clusters can take time to become available, especially during startup
	for attempt := 1; attempt <= cfg.MaxRetryAttempts; attempt++ {
		log.Printf("Attempting to connect to Cassandra (attempt %d/%d)", attempt, cfg.MaxRetryAttempts)

		// Create cluster configuration for this attempt
		cluster := gocql.NewCluster(cfg.Hosts...)

		// Configure authentication if credentials are provided
		if cfg.Username != "" && cfg.Password != "" {
			cluster.Authenticator = gocql.PasswordAuthenticator{
				Username: cfg.Username,
				Password: cfg.Password,
			}
		}

		// Set timeouts for better reliability in production environments
		cluster.ConnectTimeout = cfg.ConnectTimeout
		cluster.Timeout = cfg.RequestTimeout

		// Configure consistency level - this affects read/write guarantees
		consistency, parseErr := parseConsistency(cfg.Consistency)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid consistency level: %w", parseErr)
		}
		cluster.Consistency = consistency

		// CRITICAL: Set up token-aware host policy for optimal performance
		// This ensures queries are routed directly to nodes that own the data,
		// reducing network hops and dramatically improving response times
		cluster.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(gocql.RoundRobinHostPolicy())

		// Configure connection pooling for better throughput
		cluster.NumConns = 2
		cluster.MaxPreparedStmts = 1000
		cluster.MaxRoutingKeyInfo = 1000

		// First, try to connect without specifying a keyspace
		// This allows us to create the keyspace if it doesn't exist
		session, err = cluster.CreateSession()
		if err == nil {
			// Connection successful! Now create keyspace with configurable replication
			if createErr := createKeyspaceIfNotExists(session, cfg.Keyspace, cfg.ReplicationFactor); createErr != nil {
				session.Close()
				return nil, fmt.Errorf("failed to create keyspace: %w", createErr)
			}

			// Close the session and reconnect with the keyspace specified
			session.Close()
			cluster.Keyspace = cfg.Keyspace
			session, err = cluster.CreateSession()
			if err == nil {
				// Final verification - ping the database to ensure everything works
				if pingErr := pingDatabase(session); pingErr != nil {
					session.Close()
					err = fmt.Errorf("database ping failed: %w", pingErr)
				} else {
					// Success! Break out of retry loop
					log.Printf("Successfully connected to Cassandra cluster")
					break
				}
			}
		}

		// Connection failed, wait before retrying (unless this was the last attempt)
		if attempt < cfg.MaxRetryAttempts {
			log.Printf("Connection attempt %d failed: %v. Retrying in %v...", attempt, err, cfg.RetryDelay)
			time.Sleep(cfg.RetryDelay)
		}
	}

	// If we've exhausted all retry attempts, return the error
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Cassandra after %d attempts: %w", cfg.MaxRetryAttempts, err)
	}

	return &CassandraDB{
		Session: session,
		config:  cfg,
	}, nil
}

// createKeyspaceIfNotExists creates the application keyspace with specified replication factor
// This approach allows for environment-specific replication configuration
func createKeyspaceIfNotExists(session *gocql.Session, keyspace string, replicationFactor int) error {
	// Build the CREATE KEYSPACE query with configurable replication
	// SimpleStrategy is appropriate for single-datacenter deployments
	// For multi-datacenter setups, you would use NetworkTopologyStrategy
	query := fmt.Sprintf(`
        CREATE KEYSPACE IF NOT EXISTS %s
        WITH REPLICATION = {
            'class': 'SimpleStrategy',
            'replication_factor': %d
        }`, keyspace, replicationFactor)

	if err := session.Query(query).Exec(); err != nil {
		return fmt.Errorf("failed to create keyspace %s with replication factor %d: %w", keyspace, replicationFactor, err)
	}

	log.Printf("Keyspace %s configured with replication factor %d", keyspace, replicationFactor)
	return nil
}

// pingDatabase verifies that we can successfully interact with the database
// This is a sanity check to ensure the connection is truly functional
func pingDatabase(session *gocql.Session) error {
	var result string

	// Execute a simple query that should always work
	if err := session.Query("SELECT uuid() FROM system.local").Scan(&result); err != nil {
		return fmt.Errorf("ping query failed: %w", err)
	}

	// Verify we got a meaningful response
	if result == "" {
		return fmt.Errorf("ping query returned empty result")
	}

	return nil
}

// Close terminates the database connection
func (db *CassandraDB) Close() {
	if db.Session != nil {
		db.Session.Close()
	}
}

// Health checks if the database connection is still alive
// This is useful for health check endpoints
func (db *CassandraDB) Health() error {
	return db.Session.Query("SELECT now() FROM system.local").Exec()
}

// parseConsistency converts string consistency level to gocql constant
func parseConsistency(consistency string) (gocql.Consistency, error) {
	switch consistency {
	case "any":
		return gocql.Any, nil
	case "one":
		return gocql.One, nil
	case "two":
		return gocql.Two, nil
	case "three":
		return gocql.Three, nil
	case "quorum":
		return gocql.Quorum, nil
	case "all":
		return gocql.All, nil
	case "local_quorum":
		return gocql.LocalQuorum, nil
	case "each_quorum":
		return gocql.EachQuorum, nil
	case "local_one":
		return gocql.LocalOne, nil
	default:
		return gocql.Quorum, fmt.Errorf("unknown consistency level: %s", consistency)
	}
}
