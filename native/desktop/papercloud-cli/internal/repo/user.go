package repo

import (
	"context"
	"fmt"
	"log/slog"

	domain "github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/domain/user"
	disk "github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/pkg/storage"
)

type UserRepo struct {
	logger   *slog.Logger
	dbClient disk.Storage
}

func NewUserRepo(logger *slog.Logger, db disk.Storage) domain.Repository {
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
			slog.Any("email", email),
			slog.String("bin", string(bBytes)),
			slog.Any("error", err))
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
				slog.String("key", string(key)),
				slog.String("value", string(value)),
				slog.Any("error", err))
			return err
		}

		res = append(res, account)

		// Return nil to indicate success
		return nil
	})

	return res, err
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
