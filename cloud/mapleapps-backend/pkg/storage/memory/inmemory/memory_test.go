package inmemory

import (
	"reflect"
	"testing"

	"go.uber.org/zap"
)

// TestNewInMemoryStorage verifies that the NewInMemoryStorage function
// correctly initializes a new storage instance
func TestNewInMemoryStorage(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	storage := NewInMemoryStorage(logger)

	if storage == nil {
		t.Fatal("Expected non-nil storage instance")
	}

	// Type assertion to verify we get the correct implementation
	_, ok := storage.(*keyValueStorerImpl)
	if !ok {
		t.Fatal("Expected keyValueStorerImpl instance")
	}
}

// TestBasicOperations tests the basic Set/Get/Delete operations
func TestBasicOperations(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	storage := NewInMemoryStorage(logger)

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
		_, err := storage.Get("non-existent")
		if err == nil {
			t.Error("Expected error for non-existent key")
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
		_, err = storage.Get(key)
		if err == nil {
			t.Error("Expected error after deletion")
		}
	})
}

// TestIteration tests the Iterate functionality
func TestIteration(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	storage := NewInMemoryStorage(logger)

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
			found[string(key)] = value
			return nil
		})

		if err != nil {
			t.Fatalf("Iteration failed: %v", err)
		}

		if !reflect.DeepEqual(testData, found) {
			t.Errorf("Iteration results don't match: got %v, want %v", found, testData)
		}
	})

	// Test filtered iteration
	t.Run("Filtered Iteration", func(t *testing.T) {
		filterKeys := []string{"key1", "key3"}
		found := make(map[string][]byte)

		err := storage.IterateWithFilterByKeys(filterKeys, func(key, value []byte) error {
			found[string(key)] = value
			return nil
		})

		if err != nil {
			t.Fatalf("Filtered iteration failed: %v", err)
		}

		// Verify only requested keys were returned
		if len(found) != len(filterKeys) {
			t.Errorf("Expected %d items, got %d", len(filterKeys), len(found))
		}

		for _, k := range filterKeys {
			if !reflect.DeepEqual(found[k], testData[k]) {
				t.Errorf("Filtered data mismatch for key %s: got %v, want %v", k, found[k], testData[k])
			}
		}
	})
}

// TestTransactions tests the transaction-related functionality
func TestTransactions(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	storage := NewInMemoryStorage(logger)

	// Test basic transaction commit
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

	// Test transaction discard
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
		_, err = storage.Get(key)
		if err == nil {
			t.Error("Expected error getting discarded value")
		}
	})

	// Test transaction behavior with multiple opens
	t.Run("Multiple Transaction Opens", func(t *testing.T) {
		// Set initial value
		err := storage.Set("tx-test", []byte("initial"))
		if err != nil {
			t.Fatalf("Failed to set initial value: %v", err)
		}

		// First transaction
		err = storage.OpenTransaction()
		if err != nil {
			t.Fatalf("Failed to open first transaction: %v", err)
		}

		// Modify value
		err = storage.Set("tx-test", []byte("modified"))
		if err != nil {
			t.Fatalf("Failed to set value in transaction: %v", err)
		}

		// Opening another transaction while one is in progress overwrites the transaction data
		err = storage.OpenTransaction()
		if err != nil {
			t.Fatalf("Failed to open second transaction: %v", err)
		}

		// Modify value again
		err = storage.Set("tx-test", []byte("final"))
		if err != nil {
			t.Fatalf("Failed to set value in second transaction: %v", err)
		}

		// Commit the transaction (only need to commit once as there's only one transaction state)
		err = storage.CommitTransaction()
		if err != nil {
			t.Fatalf("Failed to commit transaction: %v", err)
		}

		// Verify attempting to commit again fails since transaction state is cleared
		err = storage.CommitTransaction()
		if err == nil {
			t.Error("Expected error when committing with no transaction in progress")
		}

		// Verify final value
		val, err := storage.Get("tx-test")
		if err != nil {
			t.Fatalf("Failed to get final value: %v", err)
		}

		if !reflect.DeepEqual(val, []byte("final")) {
			t.Errorf("Unexpected final value: got %s, want %s", string(val), "final")
		}
	})
}

// TestClose verifies the Close functionality
func TestClose(t *testing.T) {

	logger, _ := zap.NewDevelopment()
	storage := NewInMemoryStorage(logger)

	// Add some data
	err := storage.Set("test", []byte("value"))
	if err != nil {
		t.Fatalf("Failed to set test data: %v", err)
	}

	// Close storage
	err = storage.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Verify data is cleared
	_, err = storage.Get("test")
	if err == nil {
		t.Error("Expected error getting value after close")
	}
}
