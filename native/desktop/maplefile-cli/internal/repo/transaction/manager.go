// internal/repository/transaction/manager.go
package transaction

import (
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/transaction"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
)

// txManager implements the transaction.Manager interface
type txManager struct {
	userRepo       user.Repository
	collectionRepo collection.CollectionRepository
	fileRepo       file.FileRepository
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager(
	userRepo user.Repository,
	collectionRepo collection.CollectionRepository,
	fileRepo file.FileRepository,
) transaction.Manager {
	return &txManager{
		userRepo:       userRepo,
		collectionRepo: collectionRepo,
		fileRepo:       fileRepo,
	}
}

// Begin starts a new transaction
func (tm *txManager) Begin() error {
	err := tm.userRepo.OpenTransaction()
	if err != nil {
		return err
	}
	err = tm.collectionRepo.OpenTransaction()
	if err != nil {
		return err
	}
	err = tm.fileRepo.OpenTransaction()
	if err != nil {
		return err
	}
	return err
}

// Commit commits the transaction
func (tm *txManager) Commit() error {
	err := tm.userRepo.CommitTransaction()
	if err != nil {
		return err
	}
	err = tm.collectionRepo.CommitTransaction()
	if err != nil {
		return err
	}
	err = tm.fileRepo.CommitTransaction()
	if err != nil {
		return err
	}
	return err
}

// Rollback aborts the transaction
func (tm *txManager) Rollback() {
	tm.userRepo.DiscardTransaction()
	tm.collectionRepo.DiscardTransaction()
	tm.fileRepo.DiscardTransaction()
}
