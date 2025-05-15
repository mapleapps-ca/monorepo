// internal/domain/transaction/manager.go
package transaction

// Manager defines the interface for transaction management
type Manager interface {
	Begin() error
	Commit() error
	Rollback()
}
