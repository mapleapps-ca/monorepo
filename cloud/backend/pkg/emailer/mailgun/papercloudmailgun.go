package mailgun

import (
	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
)

// NewPaperCloudModuleEmailer creates a new emailer for the PaperCloud Property Evaluator module.
func NewPaperCloudModuleEmailer(cfg *config.Configuration) Emailer {
	emailerConfigProvider := NewMailgunConfigurationProvider(
		cfg.PaperCloudMailgun.SenderEmail,
		cfg.PaperCloudMailgun.Domain,
		cfg.PaperCloudMailgun.APIBase,
		cfg.PaperCloudMailgun.MaintenanceEmail,
		cfg.PaperCloudMailgun.FrontendDomain,
		cfg.PaperCloudMailgun.BackendDomain,
		cfg.PaperCloudMailgun.APIKey,
	)

	return NewEmailer(emailerConfigProvider)
}
