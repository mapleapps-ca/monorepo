package iam

import (
	"go.uber.org/fx"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/interface/http"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/repo"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/service"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/usecase"
)

func Module() fx.Option {
	return fx.Options(
		repo.Module(),
		usecase.Module(),
		service.Module(),
		http.Module(),
	)
}
