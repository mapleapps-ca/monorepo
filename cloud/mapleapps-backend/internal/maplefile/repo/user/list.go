// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/repo/user/list.go
package user

import (
	"context"

	dom_user "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/user"
)

func (s *userStorerImpl) CountByFilter(ctx context.Context, filter *dom_user.UserFilter) (uint64, error) {
	return 0, nil
}

func (impl *userStorerImpl) ListByFilter(ctx context.Context, filter *dom_user.UserFilter) (*dom_user.UserFilterResult, error) {
	return nil, nil
}
