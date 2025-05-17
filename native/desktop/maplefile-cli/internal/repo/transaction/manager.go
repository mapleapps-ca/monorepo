// internal/repository/transaction/manager.go
package transaction

import (
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localcollection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localfile"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/remotecollection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/remotefile"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/transaction"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
)

// txManager implements the transaction.Manager interface
type txManager struct {
	userRepo             user.Repository
	localcollectionRepo  localcollection.LocalCollectionRepository
	localfileRepo        localfile.LocalFileRepository
	remotecollectionRepo remotecollection.RemoteCollectionRepository
	remotefileRepo       remotefile.RemoteFileRepository
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager(
	userRepo user.Repository,
	localcollectionRepo localcollection.LocalCollectionRepository,
	localfileRepo localfile.LocalFileRepository,
	remotecollectionRepo remotecollection.RemoteCollectionRepository,
	remotefileRepo remotefile.RemoteFileRepository,
) transaction.Manager {
	return &txManager{
		userRepo:             userRepo,
		localcollectionRepo:  localcollectionRepo,
		localfileRepo:        localfileRepo,
		remotecollectionRepo: remotecollectionRepo,
		remotefileRepo:       remotefileRepo,
	}
}

// Begin starts a new transaction
func (tm *txManager) Begin() error {
	err := tm.userRepo.OpenTransaction()
	if err != nil {
		return err
	}
	err = tm.localcollectionRepo.OpenTransaction()
	if err != nil {
		return err
	}
	err = tm.localfileRepo.OpenTransaction()
	if err != nil {
		return err
	}
	err = tm.localfileRepo.OpenTransaction()
	if err != nil {
		return err
	}
	//TODO:
	// remotecollectionRepo: ,
	// remotefileRepo:       remotefileRepo,
	return err
}

// Commit commits the transaction
func (tm *txManager) Commit() error {
	err := tm.userRepo.CommitTransaction()
	if err != nil {
		return err
	}
	err = tm.localcollectionRepo.CommitTransaction()
	if err != nil {
		return err
	}
	err = tm.localfileRepo.CommitTransaction()
	if err != nil {
		return err
	}
	err = tm.localfileRepo.CommitTransaction()
	if err != nil {
		return err
	}
	//TODO:
	// remotecollectionRepo: remotecollectionRepo,
	// remotefileRepo:       remotefileRepo,
	return err
}

// Rollback aborts the transaction
func (tm *txManager) Rollback() {
	//TODO:
	tm.userRepo.DiscardTransaction()
	tm.localcollectionRepo.DiscardTransaction()
	tm.localfileRepo.DiscardTransaction()
	// remotecollectionRepo: remotecollectionRepo,
	// remotefileRepo:       remotefileRepo,
}
