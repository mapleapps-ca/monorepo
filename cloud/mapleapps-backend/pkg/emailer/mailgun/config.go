// github.com/comiccoin-network/monorepo/cloud/comiccoin/common/mailgun/config
package mailgun

type MailgunConfigurationProvider interface {
	GetSenderEmail() string
	GetDomainName() string // Deprecated
	GetBackendDomainName() string
	GetFrontendDomainName() string
	GetMaintenanceEmail() string
	GetAPIKey() string
	GetAPIBase() string
}

type mailgunConfigurationProviderImpl struct {
	senderEmail      string
	domain           string
	apiBase          string
	maintenanceEmail string
	frontendDomain   string
	backendDomain    string
	apiKey           string
}

func NewMailgunConfigurationProvider(senderEmail, domain, apiBase, maintenanceEmail, frontendDomain, backendDomain, apiKey string) MailgunConfigurationProvider {
	return &mailgunConfigurationProviderImpl{
		senderEmail:      senderEmail,
		domain:           domain,
		apiBase:          apiBase,
		maintenanceEmail: maintenanceEmail,
		frontendDomain:   frontendDomain,
		backendDomain:    backendDomain,
		apiKey:           apiKey,
	}
}

func (me *mailgunConfigurationProviderImpl) GetDomainName() string {
	return me.domain
}

func (me *mailgunConfigurationProviderImpl) GetSenderEmail() string {
	return me.senderEmail
}

func (me *mailgunConfigurationProviderImpl) GetBackendDomainName() string {
	return me.backendDomain
}

func (me *mailgunConfigurationProviderImpl) GetFrontendDomainName() string {
	return me.frontendDomain
}

func (me *mailgunConfigurationProviderImpl) GetMaintenanceEmail() string {
	return me.maintenanceEmail
}

func (me *mailgunConfigurationProviderImpl) GetAPIKey() string {
	return me.apiKey
}

func (me *mailgunConfigurationProviderImpl) GetAPIBase() string {
	return me.apiBase
}
