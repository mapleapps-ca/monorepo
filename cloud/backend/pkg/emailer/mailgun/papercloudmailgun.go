package mailgun

import (
	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
)

// NewPaperCloudPropertyEvaluatorModuleEmailer creates a new emailer for the PaperCloud Property Evaluator module.
func NewPaperCloudPropertyEvaluatorModuleEmailer(cfg *config.Configuration) Emailer {
	emailerConfigProvider := NewMailgunConfigurationProvider(
		cfg.PAPERCLOUDMailgun.SenderEmail,
		cfg.PAPERCLOUDMailgun.Domain,
		cfg.PAPERCLOUDMailgun.APIBase,
		cfg.PAPERCLOUDMailgun.MaintenanceEmail,
		cfg.PAPERCLOUDMailgun.FrontendDomain,
		cfg.PAPERCLOUDMailgun.BackendDomain,
		cfg.PAPERCLOUDMailgun.APIKey,
	)

	return NewEmailer(emailerConfigProvider)
}
