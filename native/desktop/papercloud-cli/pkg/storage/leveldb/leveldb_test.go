package leveldb

import (
	"os"
	"reflect"
	"testing"

	"go.uber.org/zap"
)

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

	logger := zap.NewExample()
	storage := NewDiskStorage(dir, "test.db", logger)

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

	impl.Close()
}

// TestBasicOperations tests the basic Set/Get/Delete operations
func TestBasicOperations(t *testing.T) {
	dir := testDir(t)
	defer cleanup(t, dir)

	storage := NewDiskStorage(dir, "test.db", zap.NewExample())
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
		if err != nil {
			t.Fatalf("Expected nil error for non-existent key, got: %v", err)
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
			t.Fatalf("Get after delete failed: %v", err)
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

	storage := NewDiskStorage(dir, "test.db", zap.NewExample())
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
			// Make a copy of the value since LevelDB reuses the byte slice
			valueCopy := make([]byte, len(value))
			copy(valueCopy, value)
			found[string(key)] = valueCopy
			return nil
		})

		if err != nil {
			t.Fatalf("Iteration failed: %v", err)
		}

		// Compare each key-value pair independently for better error reporting
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
		found := make(map[string][]byte)

		err := storage.IterateWithFilterByKeys(filterKeys, func(key, value []byte) error {
			// Make a copy of the value since LevelDB reuses the byte slice
			valueCopy := make([]byte, len(value))
			copy(valueCopy, value)
			found[string(key)] = valueCopy
			return nil
		})

		if err != nil {
			t.Fatalf("Filtered iteration failed: %v", err)
		}

		// Verify each filtered key individually
		for _, k := range filterKeys {
			expectedValue, exists := testData[k]
			if !exists {
				t.Errorf("Test data missing key %q", k)
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

	storage := NewDiskStorage(dir, "test.db", zap.NewExample())
	defer storage.Close()

	t.Run("Transaction Commit", func(t *testing.T) {
		// Start transaction
		err := storage.OpenTransaction()
		if err != nil {
			t.Fatalf("Failed to open transaction: %v", err)
		}

		// Make changes in transaction
		key := "tx-test"
		value := []byte("tx-value")

		err = storage.Set(key, value)
		if err != nil {
			t.Fatalf("Failed to set in transaction: %v", err)
		}

		// Commit transaction
		err = storage.CommitTransaction()
		if err != nil {
			t.Fatalf("Failed to commit transaction: %v", err)
		}

		// Verify changes persisted
		retrieved, err := storage.Get(key)
		if err != nil {
			t.Fatalf("Failed to get after commit: %v", err)
		}

		if !reflect.DeepEqual(retrieved, value) {
			t.Errorf("Retrieved value doesn't match after commit: got %v, want %v", retrieved, value)
		}
	})

	t.Run("Transaction Discard", func(t *testing.T) {
		// Start transaction
		err := storage.OpenTransaction()
		if err != nil {
			t.Fatalf("Failed to open transaction: %v", err)
		}

		// Make changes in transaction
		key := "discard-test"
		value := []byte("discard-value")

		err = storage.Set(key, value)
		if err != nil {
			t.Fatalf("Failed to set in transaction: %v", err)
		}

		// Discard transaction
		storage.DiscardTransaction()

		// Verify changes were not persisted
		val, err := storage.Get(key)
		if err != nil {
			t.Fatalf("Get after discard failed: %v", err)
		}
		if val != nil {
			t.Error("Expected nil value after discarding transaction")
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
		storage := NewDiskStorage(dir, dbName, zap.NewExample())
		defer storage.Close()

		err := storage.Set(key, value)
		if err != nil {
			t.Fatalf("Failed to set value: %v", err)
		}
	}()

	// Second session: verify data
	func() {
		storage := NewDiskStorage(dir, dbName, zap.NewExample())
		defer storage.Close()

		retrieved, err := storage.Get(key)
		if err != nil {
			t.Fatalf("Failed to get value: %v", err)
		}

		if !reflect.DeepEqual(retrieved, value) {
			t.Errorf("Retrieved value doesn't match after reopen: got %v, want %v", retrieved, value)
		}
	}()
}

// TestDatabaseError tests error handling for invalid database operations
func TestDatabaseError(t *testing.T) {
	dir := testDir(t)
	defer cleanup(t, dir)

	storage := NewDiskStorage(dir, "test.db", zap.NewExample())

	// Close the database to force errors
	storage.Close()

	// Test database operations with invalid state
	t.Run("Invalid Operations", func(t *testing.T) {
		// Test with empty key
		if err := storage.Set("", []byte("value")); err == nil {
			t.Error("Expected error setting empty key")
		}

		// Test with nil value
		if err := storage.Set("key", nil); err == nil {
			t.Error("Expected error setting nil value")
		}

		// Close the database
		storage.Close()

		// OpenTransaction returns nil even on error per implementation
		err := storage.OpenTransaction()
		if err != nil {
			t.Error("OpenTransaction should return nil even when db is closed")
		}

		// Verify the transaction wasn't actually created
		impl, ok := storage.(*storageImpl)
		if !ok {
			t.Fatal("Expected storageImpl instance")
		}
		if impl.transaction != nil {
			t.Error("Transaction should be nil after failed open")
		}
	})
}
