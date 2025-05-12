package inmemory

import (
	"errors"
	"fmt"
	"sync"

	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/storage"
	"go.uber.org/zap"
)

type cacheValue struct {
	value []byte
}

// keyValueStorerImpl implements the db.Database interface.
// It uses a LevelDB database to store key-value pairs.
type keyValueStorerImpl struct {
	data   map[string]cacheValue
	txData map[string]cacheValue
	lock   sync.Mutex
}

// NewInMemoryStorage creates a new instance of the keyValueStorerImpl.
func NewInMemoryStorage(logger *zap.Logger) storage.Storage {
	return &keyValueStorerImpl{
		data:   make(map[string]cacheValue),
		txData: nil,
	}
}

// Get retrieves a value from the database by its key.
// It returns an error if the key is not found.
func (impl *keyValueStorerImpl) Get(k string) ([]byte, error) {
	impl.lock.Lock()
	defer impl.lock.Unlock()

	if impl.txData != nil {
		cachedValue, ok := impl.txData[k]
		if !ok {
			return nil, fmt.Errorf("does not exist for: %v", k)
		}
		return cachedValue.value, nil
	} else {
		cachedValue, ok := impl.data[k]
		if !ok {
			return nil, fmt.Errorf("does not exist for: %v", k)
		}
		return cachedValue.value, nil
	}
}

// Set sets a value in the database by its key.
// It returns an error if the operation fails.
func (impl *keyValueStorerImpl) Set(k string, val []byte) error {
	impl.lock.Lock()
	defer impl.lock.Unlock()

	if impl.txData != nil {
		impl.txData[k] = cacheValue{
			value: val,
		}
	} else {
		impl.data[k] = cacheValue{
			value: val,
		}
	}
	return nil
}

// Delete deletes a value from the database by its key.
// It returns an error if the operation fails.
func (impl *keyValueStorerImpl) Delete(k string) error {
	impl.lock.Lock()
	defer impl.lock.Unlock()

	if impl.txData != nil {
		delete(impl.txData, k)
	} else {
		delete(impl.data, k)
	}
	return nil
}

// Iterate iterates over the key-value pairs in the database, starting from the specified key prefix.
// It calls the provided function for each pair.
// It returns an error if the iteration fails.
func (impl *keyValueStorerImpl) Iterate(processFunc func(key, value []byte) error) error {
	impl.lock.Lock()
	defer impl.lock.Unlock()

	if impl.txData != nil {
		// Iterate over the key-value pairs in the database, starting from the starting point
		for k, v := range impl.txData {
			// Call the provided function for each pair
			if err := processFunc([]byte(k), v.value); err != nil {
				return err
			}
		}
	} else {
		// Iterate over the key-value pairs in the database, starting from the starting point
		for k, v := range impl.data {
			// Call the provided function for each pair
			if err := processFunc([]byte(k), v.value); err != nil {
				return err
			}
		}
	}

	return nil
}

func (impl *keyValueStorerImpl) IterateWithFilterByKeys(ks []string, processFunc func(key, value []byte) error) error {
	impl.lock.Lock()
	defer impl.lock.Unlock()

	if impl.txData != nil {
		// Iterate over the key-value pairs in the database, starting from the starting point
		for k, v := range impl.txData {
			// Iterate over our keys to search by.
			for _, searchK := range ks {
				// If the item we currently have matches our keys then execute.
				if k == searchK {
					// Call the provided function for each pair
					if err := processFunc([]byte(k), v.value); err != nil {
						return err
					}
				}
			}

		}
	} else {
		// Iterate over the key-value pairs in the database, starting from the starting point
		for k, v := range impl.data {
			// Iterate over our keys to search by.
			for _, searchK := range ks {
				// If the item we currently have matches our keys then execute.
				if k == searchK {
					// Call the provided function for each pair
					if err := processFunc([]byte(k), v.value); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

// Close closes the database.
// It returns an error if the operation fails.
func (impl *keyValueStorerImpl) Close() error {
	impl.lock.Lock()
	defer impl.lock.Unlock()

	// Clear the data map
	impl.data = make(map[string]cacheValue)

	return nil
}

func (impl *keyValueStorerImpl) OpenTransaction() error {
	impl.lock.Lock()
	defer impl.lock.Unlock()

	// Create a new transaction by creating a copy of the current data
	impl.txData = make(map[string]cacheValue)
	for k, v := range impl.data {
		impl.txData[k] = v
	}

	return nil
}

func (impl *keyValueStorerImpl) CommitTransaction() error {
	impl.lock.Lock()
	defer impl.lock.Unlock()

	// Check if a transaction is in progress
	if impl.txData == nil {
		return errors.New("no transaction in progress")
	}

	// Update the current data with the transaction data
	impl.data = impl.txData
	impl.txData = nil

	return nil
}

func (impl *keyValueStorerImpl) DiscardTransaction() {
	impl.lock.Lock()
	defer impl.lock.Unlock()

	// Check if a transaction is in progress
	if impl.txData != nil {
		impl.txData = nil
	}
}
