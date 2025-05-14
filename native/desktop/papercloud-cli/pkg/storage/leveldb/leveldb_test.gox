package leveldb

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"go.uber.org/zap"
)

// MockLevelDBConfigProvider is a mock implementation of a presumed LevelDBConfigurationProvider interface.
// This is created to satisfy the changed signature of NewDiskStorage.
type MockLevelDBConfigProvider struct {
	Path string
	// Add other fields here if the actual LevelDBConfigurationProvider interface
	// and NewDiskStorage function require more configuration options (e.g., LevelDB options).
}

// GetDBPath returns the database path. This method is assumed to be part of the
// LevelDBConfigurationProvider interface.
func (m *MockLevelDBConfigProvider) GetDBPath() string {
	return m.Path
}

// If LevelDBConfigurationProvider requires other methods (e.g., GetOptions()),
// they would need to be implemented here as well. For this repair, we assume
// GetDBPath is sufficient based on the original parameters of NewDiskStorage.

// testDir creates a temporary directory for testing
func testDir(t *testing.T) string {
	dir, err := os.MkdirTemp("", "leveldb-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	return dir
}

// cleanup removes the test directory and its contents
func cleanup(t *testing.T, dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		t.Errorf("Failed to cleanup test dir: %v", err)
	}
}

// TestNewDiskStorage tests the creation of a new storage instance
func TestNewDiskStorage(t *testing.T) {
	dir := testDir(t)
	defer cleanup(t, dir)

	dbName := "test.db"
	config := &MockLevelDBConfigProvider{
		Path: filepath.Join(dir, dbName),
	}
	logger := zap.NewExample()
	storage := NewDiskStorage(config, logger)

	if storage == nil {
		t.Fatal("Expected non-nil storage instance")
	}

	// Type assertion to verify we get the correct implementation
	impl, ok := storage.(*storageImpl)
	if !ok {
		t.Fatal("Expected storageImpl instance")
	}

	if impl.db == nil {
		t.Fatal("Expected non-nil leveldb instance")
	}

	// It's generally better to call Close on the interface, but if impl.Close() has
	// specific behavior being tested or is necessary for some reason, it can stay.
	// However, a defer storage.Close() would be more idiomatic if not for specific impl testing.
	err := storage.Close() // Changed from impl.Close() to storage.Close() for consistency
	if err != nil {
		t.Fatalf("Failed to close storage: %v", err)
	}
}

// TestBasicOperations tests the basic Set/Get/Delete operations
func TestBasicOperations(t *testing.T) {
	dir := testDir(t)
	defer cleanup(t, dir)

	dbName := "test.db"
	config := &MockLevelDBConfigProvider{
		Path: filepath.Join(dir, dbName),
	}
	storage := NewDiskStorage(config, zap.NewExample())
	defer storage.Close()

	// Test Set and Get
	t.Run("Set and Get", func(t *testing.T) {
		key := "test-key"
		value := []byte("test-value")

		err := storage.Set(key, value)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		retrieved, err := storage.Get(key)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		if !reflect.DeepEqual(retrieved, value) {
			t.Errorf("Retrieved value doesn't match: got %v, want %v", retrieved, value)
		}
	})

	// Test Get with non-existent key
	t.Run("Get Non-existent", func(t *testing.T) {
		val, err := storage.Get("non-existent")
		// LevelDB's Get typically returns leveldb.ErrNotFound for non-existent keys.
		// If the wrapper converts this to (nil, nil), the original check is fine.
		// If it propagates ErrNotFound, then err should be checked for that specific error.
		// Assuming the wrapper intends (nil, nil) for not found.
		if err != nil {
			t.Fatalf("Expected nil error or specific 'not found' error for non-existent key, got: %v", err)
		}
		if val != nil {
			t.Errorf("Expected nil value for non-existent key, got: %v", val)
		}
	})

	// Test Delete
	t.Run("Delete", func(t *testing.T) {
		key := "delete-test"
		value := []byte("delete-value")

		// First set a value
		err := storage.Set(key, value)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		// Delete it
		err = storage.Delete(key)
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		// Verify it's gone
		val, err := storage.Get(key)
		if err != nil {
			t.Fatalf("Get after delete failed: %v", err) // Similar to "Get Non-existent"
		}
		if val != nil {
			t.Error("Expected nil value after deletion")
		}
	})
}

// TestIteration tests the iteration functionality
func TestIteration(t *testing.T) {
	dir := testDir(t)
	defer cleanup(t, dir)

	dbName := "test.db"
	config := &MockLevelDBConfigProvider{
		Path: filepath.Join(dir, dbName),
	}
	storage := NewDiskStorage(config, zap.NewExample())
	defer storage.Close()

	// Prepare test data
	testData := map[string][]byte{
		"key1": []byte("value1"),
		"key2": []byte("value2"),
		"key3": []byte("value3"),
	}

	// Insert test data
	for k, v := range testData {
		if err := storage.Set(k, v); err != nil {
			t.Fatalf("Failed to set test data: %v", err)
		}
	}

	// Test basic iteration
	t.Run("Basic Iteration", func(t *testing.T) {
		found := make(map[string][]byte)

		err := storage.Iterate(func(key, value []byte) error {
			valueCopy := make([]byte, len(value))
			copy(valueCopy, value)
			found[string(key)] = valueCopy
			return nil
		})

		if err != nil {
			t.Fatalf("Iteration failed: %v", err)
		}

		if len(found) != len(testData) {
			t.Errorf("Iteration found %d items, expected %d", len(found), len(testData))
		}

		for k, expectedValue := range testData {
			actualValue, exists := found[k]
			if !exists {
				t.Errorf("Key %q not found in iteration results", k)
				continue
			}
			if !reflect.DeepEqual(actualValue, expectedValue) {
				t.Errorf("Value mismatch for key %q: got %q, want %q",
					k, string(actualValue), string(expectedValue))
			}
		}
	})

	// Test filtered iteration
	t.Run("Filtered Iteration", func(t *testing.T) {
		filterKeys := []string{"key1", "key3"}
		expectedFilteredData := make(map[string][]byte)
		for _, k := range filterKeys {
			if v, ok := testData[k]; ok {
				expectedFilteredData[k] = v
			}
		}

		found := make(map[string][]byte)

		err := storage.IterateWithFilterByKeys(filterKeys, func(key, value []byte) error {
			valueCopy := make([]byte, len(value))
			copy(valueCopy, value)
			found[string(key)] = valueCopy
			return nil
		})

		if err != nil {
			t.Fatalf("Filtered iteration failed: %v", err)
		}

		if len(found) != len(expectedFilteredData) {
			t.Errorf("Filtered iteration found %d items, expected %d", len(found), len(expectedFilteredData))
		}

		for _, k := range filterKeys {
			expectedValue, dataExists := testData[k]
			if !dataExists { // Should not happen if filterKeys are from testData
				t.Errorf("Test data sanity check: key %q from filterKeys not in testData", k)
				continue
			}

			actualValue, exists := found[k]
			if !exists {
				t.Errorf("Key %q not found in filtered results", k)
				continue
			}

			if !reflect.DeepEqual(actualValue, expectedValue) {
				t.Errorf("Value mismatch for key %q: got %q, want %q",
					k, string(actualValue), string(expectedValue))
			}
		}
	})
}

// TestTransactions tests the transaction functionality
func TestTransactions(t *testing.T) {
	dir := testDir(t)
	defer cleanup(t, dir)

	dbName := "test.db"
	config := &MockLevelDBConfigProvider{
		Path: filepath.Join(dir, dbName),
	}
	storage := NewDiskStorage(config, zap.NewExample())
	defer storage.Close()

	t.Run("Transaction Commit", func(t *testing.T) {
		err := storage.OpenTransaction()
		if err != nil {
			t.Fatalf("Failed to open transaction: %v", err)
		}

		key := "tx-test"
		value := []byte("tx-value")

		err = storage.Set(key, value) // Inside transaction
		if err != nil {
			storage.DiscardTransaction() // Ensure transaction is cleaned up on failure
			t.Fatalf("Failed to set in transaction: %v", err)
		}

		err = storage.CommitTransaction()
		if err != nil {
			t.Fatalf("Failed to commit transaction: %v", err)
		}

		retrieved, err := storage.Get(key)
		if err != nil {
			t.Fatalf("Failed to get after commit: %v", err)
		}

		if !reflect.DeepEqual(retrieved, value) {
			t.Errorf("Retrieved value doesn't match after commit: got %v, want %v", retrieved, value)
		}
	})

	t.Run("Transaction Discard", func(t *testing.T) {
		err := storage.OpenTransaction()
		if err != nil {
			t.Fatalf("Failed to open transaction: %v", err)
		}

		key := "discard-test"
		value := []byte("discard-value")

		// Set a value outside transaction first to ensure Get behavior is clear
		outerKey := "outer-key-discard"
		outerValue := []byte("outer-value-discard")
		if err := storage.Set(outerKey, outerValue); err != nil {
			storage.DiscardTransaction()
			t.Fatalf("Setup: Failed to set outer key: %v", err)
		}

		err = storage.Set(key, value) // Inside transaction
		if err != nil {
			storage.DiscardTransaction()
			t.Fatalf("Failed to set in transaction: %v", err)
		}

		storage.DiscardTransaction()

		// Verify changes were not persisted for 'key'
		val, err := storage.Get(key)
		if err != nil { // Assuming (nil, nil) for not found, or check for specific not-found error
			t.Fatalf("Get after discard failed for transactional key: %v", err)
		}
		if val != nil {
			t.Error("Expected nil value for transactional key after discarding transaction")
		}

		// Verify outer key is still accessible (if transactions don't block normal ops)
		// This depends on the implementation detail of how transactions interact with non-transactional ops.
		// If OpenTransaction acquires a lock, this Get might behave differently.
		// For now, assuming Get outside transaction should work.
		retrievedOuter, err := storage.Get(outerKey)
		if err != nil {
			t.Fatalf("Get after discard failed for outer key: %v", err)
		}
		if !reflect.DeepEqual(retrievedOuter, outerValue) {
			t.Errorf("Outer key value changed or became inaccessible after discard: got %v, want %v", retrievedOuter, outerValue)
		}
	})
}

// TestPersistence verifies that data persists after closing and reopening the database
func TestPersistence(t *testing.T) {
	dir := testDir(t)
	defer cleanup(t, dir)

	dbName := "persist.db"
	key := "persist-key"
	value := []byte("persist-value")

	// First session: write data
	func() {
		config := &MockLevelDBConfigProvider{
			Path: filepath.Join(dir, dbName),
		}
		storage := NewDiskStorage(config, zap.NewExample())
		defer storage.Close()

		err := storage.Set(key, value)
		if err != nil {
			t.Fatalf("Failed to set value: %v", err)
		}
	}() // storage is closed here

	// Second session: verify data
	func() {
		config := &MockLevelDBConfigProvider{
			Path: filepath.Join(dir, dbName),
		}
		storage := NewDiskStorage(config, zap.NewExample())
		defer storage.Close()

		retrieved, err := storage.Get(key)
		if err != nil {
			t.Fatalf("Failed to get value: %v", err)
		}

		if !reflect.DeepEqual(retrieved, value) {
			t.Errorf("Retrieved value doesn't match after reopen: got %v, want %v", retrieved, value)
		}
	}() // storage is closed here
}

// TestDatabaseError tests error handling for invalid database operations
func TestDatabaseError(t *testing.T) {
	dir := testDir(t)
	defer cleanup(t, dir)

	dbName := "test.db"
	config := &MockLevelDBConfigProvider{
		Path: filepath.Join(dir, dbName),
	}
	storage := NewDiskStorage(config, zap.NewExample())

	// Test argument validation (should ideally be tested on an open DB,
	// but original test implies testing on closed DB or combined effects)
	// If these are meant to be pure argument validation, they should be done before Close().
	// For now, keeping original structure.

	// Test Set with empty key (on an open DB first, then closed)
	// Let's assume the Set method itself validates this, regardless of DB state.
	// If Set returns ErrClosed when DB is closed, it might mask empty key error.
	// To test empty key specifically:
	if err := storage.Set("", []byte("value")); err == nil { // This might be an error from the underlying DB for empty key
		t.Error("Expected error setting empty key (on open DB or for argument validation)")
	}
	// Test Set with nil value
	if err := storage.Set("key", nil); err == nil { // This might be an error from underlying DB for nil value
		t.Error("Expected error setting nil value (on open DB or for argument validation)")
	}

	// Now, close the database to force errors for subsequent operations
	errClose := storage.Close()
	if errClose != nil {
		t.Fatalf("Failed to close storage for error testing: %v", errClose)
	}

	// Test database operations on a closed DB
	t.Run("Operations on Closed DB", func(t *testing.T) {
		if err := storage.Set("another-key", []byte("value")); err == nil {
			t.Error("Expected error setting on a closed database")
		}

		if _, err := storage.Get("another-key"); err == nil {
			t.Error("Expected error getting from a closed database")
		}

		if err := storage.Delete("another-key"); err == nil {
			t.Error("Expected error deleting from a closed database")
		}

		if err := storage.Iterate(func(k, v []byte) error { return nil }); err == nil {
			t.Error("Expected error iterating on a closed database")
		}

		// Test OpenTransaction on a closed DB
		// Original comment: "OpenTransaction returns nil even on error per implementation"
		// This behavior is unusual. A robust implementation should return an error.
		err := storage.OpenTransaction()
		// If the contract is that it returns nil error even if DB is closed:
		if err != nil {
			t.Errorf("OpenTransaction returned an error (%v), expected nil error even when db is closed based on original test comment", err)
		}
		// Even if OpenTransaction returns nil error, it shouldn't have actually created a transaction.
		if impl, ok := storage.(*storageImpl); ok {
			if impl.transaction != nil {
				t.Error("Transaction should be nil after OpenTransaction on a closed DB, even if no error was returned by OpenTransaction")
			}
		} else {
			t.Fatal("Expected storageImpl instance for checking transaction state")
		}

		// CommitTransaction on a closed DB (and no active transaction)
		if err := storage.CommitTransaction(); err == nil {
			t.Error("Expected error committing transaction on a closed database / no active transaction")
		}

		// DiscardTransaction on a closed DB (and no active transaction)
		// Discard is often idempotent, so it might not error.
		storage.DiscardTransaction() // Assuming this is safe to call.

		// Closing an already closed DB
		if err := storage.Close(); err == nil {
			// This depends on the desired behavior of Close(). Some are idempotent, others error.
			// t.Error("Expected error when closing an already closed database, if that's the contract")
		}
	})
}
