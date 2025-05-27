// internal/repo/transaction/manager.go
package transaction

import (
	"sync"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/transaction"
)

// transactionManager implements the transaction.Manager interface
type transactionManager struct {
	logger             *zap.Logger
	collectionRepo     collection.CollectionRepository
	fileRepo           file.FileRepository
	inTransaction      bool
	transactionStarted bool
	mutex              sync.Mutex
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager(
	logger *zap.Logger,
	collectionRepo collection.CollectionRepository,
	fileRepo file.FileRepository,
) transaction.Manager {
	return &transactionManager{
		logger:         logger,
		collectionRepo: collectionRepo,
		fileRepo:       fileRepo,
		inTransaction:  false,
		mutex:          sync.Mutex{},
	}
}

// Begin starts a new transaction across all repositories
func (tm *transactionManager) Begin() error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if tm.inTransaction {
		tm.logger.Warn("Transaction already in progress")
		return errors.NewAppError("transaction already in progress", nil)
	}

	tm.logger.Debug("Beginning transaction")

	// Begin transaction on collection repository
	if err := tm.collectionRepo.OpenTransaction(); err != nil {
		tm.logger.Error("Failed to begin transaction on collection repository", zap.Error(err))
		return errors.NewAppError("failed to begin collection transaction", err)
	}

	// Begin transaction on file repository
	if err := tm.fileRepo.OpenTransaction(); err != nil {
		tm.logger.Error("Failed to begin transaction on file repository", zap.Error(err))
		// Rollback collection transaction
		tm.collectionRepo.DiscardTransaction()
		return errors.NewAppError("failed to begin file transaction", err)
	}

	tm.inTransaction = true
	tm.transactionStarted = true
	tm.logger.Debug("Transaction started successfully")
	return nil
}

// Commit commits the transaction across all repositories
func (tm *transactionManager) Commit() error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if !tm.inTransaction {
		tm.logger.Warn("No transaction in progress to commit")
		return errors.NewAppError("no transaction in progress", nil)
	}

	tm.logger.Debug("Committing transaction")

	var commitErrors []error

	// Commit collection repository transaction
	if err := tm.collectionRepo.CommitTransaction(); err != nil {
		tm.logger.Error("Failed to commit collection transaction", zap.Error(err))
		commitErrors = append(commitErrors, err)
	}

	// Commit file repository transaction
	if err := tm.fileRepo.CommitTransaction(); err != nil {
		tm.logger.Error("Failed to commit file transaction", zap.Error(err))
		commitErrors = append(commitErrors, err)
	}

	tm.inTransaction = false
	tm.transactionStarted = false

	if len(commitErrors) > 0 {
		tm.logger.Error("Transaction commit completed with errors", zap.Int("errorCount", len(commitErrors)))
		return errors.NewAppError("transaction commit failed", commitErrors[0])
	}

	tm.logger.Debug("Transaction committed successfully")
	return nil
}

// Rollback rolls back the transaction across all repositories
func (tm *transactionManager) Rollback() error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if !tm.inTransaction {
		tm.logger.Warn("No transaction in progress to rollback")
		return nil // Not an error, just a no-op
	}

	tm.logger.Debug("Rolling back transaction")

	// Discard collection repository transaction
	tm.collectionRepo.DiscardTransaction()

	// Discard file repository transaction
	tm.fileRepo.DiscardTransaction()

	tm.inTransaction = false
	tm.transactionStarted = false

	tm.logger.Debug("Transaction rolled back successfully")
	return nil
}

// IsInTransaction returns true if currently in a transaction
func (tm *transactionManager) IsInTransaction() bool {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	return tm.inTransaction
}
