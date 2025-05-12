package storage

// Storage interface defines the methods that can be used to interact with a key-value database.
type Storage interface {
	// Get returns the value associated with the specified key, or an error if the key is not found.
	Get(key string) ([]byte, error)

	// Set sets the value associated with the specified key.
	// If the key already exists, its value is updated.
	Set(key string, val []byte) error

	// Delete removes the value associated with the specified key from the database.
	Delete(key string) error

	// Iterate is similar to View, but allows the iteration to start from a specific key prefix.
	// The seekThenIterateKey parameter can be used to specify a key to seek to before starting the iteration.
	Iterate(processFunc func(key, value []byte) error) error

	IterateWithFilterByKeys(ks []string, processFunc func(key, value []byte) error) error

	// Close closes the database, releasing any system resources it holds.
	Close() error

	OpenTransaction() error

	CommitTransaction() error

	DiscardTransaction()
}
