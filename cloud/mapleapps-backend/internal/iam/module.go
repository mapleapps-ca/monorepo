package iam

import (
	"go.uber.org/fx"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/repo"
	// "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/interface/http"//TODO: Implement
	// "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/service" //TODO: Implement
	// "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/usecase" //TODO: Implement
)

func Module() fx.Option {
	return fx.Options(
		repo.Module(),
		// usecase.Module(), //TODO: Implement
		// service.Module(), //TODO: Implement
		// http.Module(),//TODO: Implement
	)
}
