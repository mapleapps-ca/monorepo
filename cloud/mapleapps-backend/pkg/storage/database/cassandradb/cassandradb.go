package cassandradb

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	c "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
)

func CreateSessionWithRetryAndKeyspace(consistency gocql.Consistency, keyspace string, replicationFactor int64, hosts ...string) (*gocql.Session, error) {
	var session *gocql.Session
	var err error

	for attempt := 1; attempt <= 20; attempt++ {
		cluster := gocql.NewCluster(hosts...)
		cluster.Timeout = 25 * time.Second
		cluster.Consistency = consistency

		// Set up token-aware host policy for the cluster
		cluster.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(gocql.RoundRobinHostPolicy())

		// Attempt to create a session
		session, err = cluster.CreateSession()
		if err == nil {
			// DEVELOPERS NOTE:
			// If no errors occured then we can proceed with the following two
			// steps before exiting our provider.

			//TODO: Alter replication factor on starup

			// STEP 1 OF 2:
			// Create the keyspace if we haven't before.
			query := fmt.Sprintf("CREATE KEYSPACE IF NOT EXISTS %v WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor': %v}", keyspace, replicationFactor)
			if err := session.Query(query).Exec(); err != nil {
				return nil, err
			}

			// STEP 2 OF 2:
			// Successfully connected, break out of the loop
			break
		}

		// Wait for a short duration before the next attempt
		time.Sleep(25 * time.Second)
	}

	if err != nil {
		// Failed to connect after 10 attempts, generate panic
		panic(fmt.Sprintf("Failed to connect to the database after 10 attempts: %v", err))
	}

	return session, nil
}

func createKeyspaceIfNotExists(keyspace string, replicationFactor int64, hosts ...string) error {
	cluster := gocql.NewCluster(hosts...)
	session, err := cluster.CreateSession()
	if err != nil {
		return err
	}
	defer session.Close()

	query := fmt.Sprintf("CREATE KEYSPACE IF NOT EXISTS %v WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor': %v}", keyspace, replicationFactor)
	if err := session.Query(query).Exec(); err != nil {
		return err
	}

	return nil
}

func ping(session *gocql.Session) error {
	var str = new(string)
	if err := session.Query("SELECT uuid() FROM system.local;").Scan(str); err != nil {
		return err
	}
	if str == nil || len(*str) == 0 {
		return errors.New("failed sanity check")
	}
	return nil
}

// NewProvider function will initiate a connection with Cassandra DB and supply an instance of the connection to our app. This code is useful for dependency injection.
func NewProvider(appCfg *c.Configuration, logger *zap.Logger) *gocql.Session {
	logger.Info("cassandra storage initializing...")
	session, err := CreateSessionWithRetryAndKeyspace(gocql.Quorum, "mapleapps", appCfg.DB.KeyspaceReplicationFactor, appCfg.DB.Hosts...)
	if err != nil {
		log.Fatal(err)
	}

	// Confirm we are able to interact with the database cluster before
	// proceeding further. If the ping fails then we have not successfully
	// connected with the Cassandra.
	if err := ping(session); err != nil {
		log.Fatal(err)
	}

	logger.Info("cassandra storage initialized successfully",
		zap.String("keyspace", "mothership"),
		zap.Any("hosts", appCfg.DB.Hosts),
		zap.Int64("replication_factor", appCfg.DB.KeyspaceReplicationFactor),
	)
	return session
}
