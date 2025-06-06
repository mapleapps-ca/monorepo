package mailgun

import (
	"context"
	"time"

	"github.com/mailgun/mailgun-go/v4"
)

type mailgunEmailer struct {
	config  MailgunConfigurationProvider
	Mailgun *mailgun.MailgunImpl
}

func NewEmailer(config MailgunConfigurationProvider) Emailer {
	// Defensive code: Make sure we have access to the file before proceeding any further with the code.
	mg := mailgun.NewMailgun(config.GetDomainName(), config.GetAPIKey())

	mg.SetAPIBase(config.GetAPIBase()) // Override to support our custom email requirements.

	return &mailgunEmailer{
		config:  config,
		Mailgun: mg,
	}
}

func (me *mailgunEmailer) Send(ctx context.Context, sender, subject, recipient, body string) error {

	message := me.Mailgun.NewMessage(sender, subject, "", recipient)
	message.SetHtml(body)

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	// Send the message with a 10 second timeout
	_, _, err := me.Mailgun.Send(ctx, message)

	if err != nil {
		return err
	}

	return nil
}

func (me *mailgunEmailer) GetDomainName() string {
	return me.config.GetDomainName()
}

func (me *mailgunEmailer) GetSenderEmail() string {
	return me.config.GetSenderEmail()
}

func (me *mailgunEmailer) GetBackendDomainName() string {
	return me.config.GetBackendDomainName()
}

func (me *mailgunEmailer) GetFrontendDomainName() string {
	return me.config.GetFrontendDomainName()
}

func (me *mailgunEmailer) GetMaintenanceEmail() string {
	return me.config.GetMaintenanceEmail()
}
