package leveldb

import (
	"log"
	"strings"

	"github.com/syndtr/goleveldb/leveldb"
	dberr "github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/pkg/storage"
)

// storageImpl implements the db.Database interface.
// It uses a LevelDB database to store key-value pairs.
type storageImpl struct {
	// The LevelDB database instance.
	db          *leveldb.DB
	transaction *leveldb.Transaction
}

// NewDiskStorage creates a new instance of the storageImpl.
// It opens the database file at the specified path and returns an error if it fails.
func NewDiskStorage(provider LevelDBConfigurationProvider, logger *zap.Logger) storage.Storage {
	if provider == nil {
		log.Fatal("NewDiskStorage: missing LevelDB configuration provider\n")
	}
	if provider.GetDBPath() == "" {
		log.Fatal("NewDiskStorage: cannot have empty filepath for the database\n")
	}
	if provider.GetDBName() == "" {
		log.Fatal("NewDiskStorage: cannot have empty db name for the database\n")
	}

	o := &opt.Options{
		Filter: filter.NewBloomFilter(10),
	}

	filePath := provider.GetDBPath() + "/" + provider.GetDBName()

	db, err := leveldb.OpenFile(filePath, o)
	if err != nil {
		log.Fatalf("NewDiskStorage: failed loading up key value storer adapter at %v with error: %v\n", filePath, err)
	}
	return &storageImpl{
		db: db,
	}
}

// Get retrieves a value from the database by its key.
// It returns an error if the key is not found.
func (impl *storageImpl) Get(k string) ([]byte, error) {
	if impl.transaction == nil {
		bin, err := impl.db.Get([]byte(k), nil)
		if err == dberr.ErrNotFound {
			return nil, nil
		}
		return bin, nil
	}

	bin, err := impl.transaction.Get([]byte(k), nil)
	if err == dberr.ErrNotFound {
		return nil, nil
	}
	return bin, nil
}

// Set sets a value in the database by its key.
// It returns an error if the operation fails.
func (impl *storageImpl) Set(k string, val []byte) error {
	if impl.transaction == nil {
		impl.db.Delete([]byte(k), nil)
		err := impl.db.Put([]byte(k), val, nil)
		if err == dberr.ErrNotFound {
			return nil
		}
		return err
	}

	impl.transaction.Delete([]byte(k), nil)
	err := impl.transaction.Put([]byte(k), val, nil)
	if err == dberr.ErrNotFound {
		return nil
	}
	return err
}

// Delete deletes a value from the database by its key.
// It returns an error if the operation fails.
func (impl *storageImpl) Delete(k string) error {
	if impl.transaction == nil {
		err := impl.db.Delete([]byte(k), nil)
		if err == dberr.ErrNotFound {
			return nil
		}
		return err
	}

	err := impl.transaction.Delete([]byte(k), nil)
	if err == dberr.ErrNotFound {
		return nil
	}
	return err
}

// Iterate iterates over the key-value pairs in the database, starting from the specified key prefix.
// It calls the provided function for each pair.
// It returns an error if the iteration fails.
func (impl *storageImpl) Iterate(processFunc func(key, value []byte) error) error {
	if impl.transaction == nil {
		iter := impl.db.NewIterator(nil, nil)
		for ok := iter.First(); ok; ok = iter.Next() {
			// Call the passed function for each key-value pair.
			err := processFunc(iter.Key(), iter.Value())
			if err == dberr.ErrNotFound {
				return nil
			}
			if err != nil {
				return err // Exit early if the processing function returns an error.
			}
		}
		iter.Release()
		return iter.Error()
	}

	iter := impl.transaction.NewIterator(nil, nil)
	for ok := iter.First(); ok; ok = iter.Next() {
		// Call the passed function for each key-value pair.
		err := processFunc(iter.Key(), iter.Value())
		if err == dberr.ErrNotFound {
			return nil
		}
		if err != nil {
			return err // Exit early if the processing function returns an error.
		}
	}
	iter.Release()
	return iter.Error()
}

func (impl *storageImpl) IterateWithFilterByKeys(ks []string, processFunc func(key, value []byte) error) error {
	if impl.transaction == nil {
		iter := impl.db.NewIterator(nil, nil)
		for ok := iter.First(); ok; ok = iter.Next() {
			// Iterate over our keys to search by.
			for _, k := range ks {
				searchKey := strings.ToLower(k)
				targetKey := strings.ToLower(string(iter.Key()))

				// If the item we currently have matches our keys then execute.
				if searchKey == targetKey {
					// Call the passed function for each key-value pair.
					err := processFunc(iter.Key(), iter.Value())
					if err == dberr.ErrNotFound {
						return nil
					}
					if err != nil {
						return err // Exit early if the processing function returns an error.
					}
				}
			}
		}
		iter.Release()
		return iter.Error()
	}

	iter := impl.transaction.NewIterator(nil, nil)
	for ok := iter.First(); ok; ok = iter.Next() {
		// Call the passed function for each key-value pair.
		err := processFunc(iter.Key(), iter.Value())
		if err == dberr.ErrNotFound {
			return nil
		}
		if err != nil {
			return err // Exit early if the processing function returns an error.
		}
	}
	iter.Release()
	return iter.Error()
}

// Close closes the database.
// It returns an error if the operation fails.
func (impl *storageImpl) Close() error {
	if impl.transaction != nil {
		impl.transaction.Discard()
	}
	return impl.db.Close()
}

func (impl *storageImpl) OpenTransaction() error {
	transaction, err := impl.db.OpenTransaction()
	if err != nil {
		return nil
	}
	impl.transaction = transaction
	return nil
}

func (impl *storageImpl) CommitTransaction() error {
	defer func() {
		impl.transaction = nil
	}()

	// Commit the snapshot to the database
	return impl.transaction.Commit()
}

func (impl *storageImpl) DiscardTransaction() {
	defer func() {
		impl.transaction = nil
	}()
	impl.transaction.Discard()
}
