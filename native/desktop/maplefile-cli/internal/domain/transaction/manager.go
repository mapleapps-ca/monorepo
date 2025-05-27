// internal/domain/transaction/interface.go
package transaction

// Manager defines the interface for transaction management
type Manager interface {
	// Begin starts a new transaction
	Begin() error

	// Commit commits the current transaction
	Commit() error

	// Rollback rolls back the current transaction
	Rollback() error

	// IsInTransaction returns true if currently in a transaction
	IsInTransaction() bool
}
