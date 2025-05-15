// Location: monorepo/native/desktop/maplefile-cli/internal/repo/user.go
package repo

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	domain "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	disk "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/storage"
)

type UserRepo struct {
	logger   *zap.Logger
	dbClient disk.Storage
}

func NewUserRepo(logger *zap.Logger, db disk.Storage) domain.Repository {
	return &UserRepo{logger, db}
}

func (r *UserRepo) UpsertByEmail(ctx context.Context, user *domain.User) error {
	bBytes, err := user.Serialize()
	if err != nil {
		return err
	}
	if err := r.dbClient.Set(fmt.Sprintf("%v", user.Email), bBytes); err != nil {
		return err
	}
	return nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	bBytes, err := r.dbClient.Get(fmt.Sprintf("%v", email))
	if err != nil {
		return nil, err
	}
	b, err := domain.NewUserFromDeserialize(bBytes)
	if err != nil {
		r.logger.Error("failed to deserialize",
			zap.Any("email", email),
			zap.String("bin", string(bBytes)),
			zap.Any("error", err))
		return nil, err
	}
	return b, nil
}

func (r *UserRepo) DeleteByEmail(ctx context.Context, email string) error {
	err := r.dbClient.Delete(fmt.Sprintf("%v", email))
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepo) ListAll(ctx context.Context) ([]*domain.User, error) {
	res := make([]*domain.User, 0)
	err := r.dbClient.Iterate(func(key, value []byte) error {
		account, err := domain.NewUserFromDeserialize(value)
		if err != nil {
			r.logger.Error("failed to deserialize",
				zap.String("key", string(key)),
				zap.String("value", string(value)),
				zap.Any("error", err))
			return err
		}

		res = append(res, account)

		// Return nil to indicate success
		return nil
	})

	return res, err
}

// UpdateVerificationStatus updates the user's verification status
func (r *UserRepo) UpdateVerificationStatus(ctx context.Context, email string, verified bool, role int8, status int8) error {
	// Get the user first
	user, err := r.GetByEmail(ctx, email)
	if err != nil {
		return errors.NewAppError("failed to get user", err)
	}

	if user == nil {
		return errors.NewAppError("user not found", nil)
	}

	// Update the user's verification status
	user.WasEmailVerified = verified
	user.Role = role
	user.Status = status
	user.ModifiedAt = time.Now()

	// Save the updated user
	return r.UpsertByEmail(ctx, user)
}

func (r *UserRepo) OpenTransaction() error {
	return r.dbClient.OpenTransaction()
}

func (r *UserRepo) CommitTransaction() error {
	return r.dbClient.CommitTransaction()
}

func (r *UserRepo) DiscardTransaction() {
	r.dbClient.DiscardTransaction()
}
