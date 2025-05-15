package mailgun

import (
	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
)

// NewMapleFileModuleEmailer creates a new emailer for the MapleFile Property Evaluator module.
func NewMapleFileModuleEmailer(cfg *config.Configuration) Emailer {
	emailerConfigProvider := NewMailgunConfigurationProvider(
		cfg.MapleFileMailgun.SenderEmail,
		cfg.MapleFileMailgun.Domain,
		cfg.MapleFileMailgun.APIBase,
		cfg.MapleFileMailgun.MaintenanceEmail,
		cfg.MapleFileMailgun.FrontendDomain,
		cfg.MapleFileMailgun.BackendDomain,
		cfg.MapleFileMailgun.APIKey,
	)

	return NewEmailer(emailerConfigProvider)
}
