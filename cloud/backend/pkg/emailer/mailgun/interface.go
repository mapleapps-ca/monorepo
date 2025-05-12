package mailgun

import "context"

type Emailer interface {
	Send(ctx context.Context, sender, subject, recipient, htmlContent string) error
	GetSenderEmail() string
	GetDomainName() string // Deprecated
	GetBackendDomainName() string
	GetFrontendDomainName() string
	GetMaintenanceEmail() string
}
