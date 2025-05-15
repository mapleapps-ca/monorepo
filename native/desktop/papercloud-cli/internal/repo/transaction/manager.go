// internal/repository/transaction/manager.go
package transaction

import (
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/domain/transaction"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/domain/user"
)

// txManager implements the transaction.Manager interface
type txManager struct {
	userRepo user.Repository
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager(userRepo user.Repository) transaction.Manager {
	return &txManager{
		userRepo: userRepo,
	}
}

// Begin starts a new transaction
func (tm *txManager) Begin() error {
	return tm.userRepo.OpenTransaction()
}

// Commit commits the transaction
func (tm *txManager) Commit() error {
	return tm.userRepo.CommitTransaction()
}

// Rollback aborts the transaction
func (tm *txManager) Rollback() {
	tm.userRepo.DiscardTransaction()
}
