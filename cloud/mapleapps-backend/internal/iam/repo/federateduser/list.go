// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/repo/federateduser/list.go
package federateduser

import (
	"context"

	dom_user "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/federateduser"
)

func (s *userStorerImpl) CountByFilter(ctx context.Context, filter *dom_user.FederatedUserFilter) (uint64, error) {
	return 0, nil
}

func (impl *userStorerImpl) ListByFilter(ctx context.Context, filter *dom_user.FederatedUserFilter) (*dom_user.FederatedUserFilterResult, error) {
	return nil, nil
}
