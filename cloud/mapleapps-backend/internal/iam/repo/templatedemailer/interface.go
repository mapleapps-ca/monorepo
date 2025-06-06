package templatedemailer

import (
	"context"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/emailer/mailgun"
)

// TemplatedEmailer Is adapter for responsive HTML email templates sender.
type TemplatedEmailer interface {
	SendUserVerificationEmail(ctx context.Context, monolithModule int, email, verificationCode, firstName string) error
	SendUserPasswordResetEmail(ctx context.Context, monolithModule int, email, verificationCode, firstName string) error
	SendUserLoginOneTimeTokenEmail(ctx context.Context, monolithModule int, email, oneTimeToken, firstName string) error
}

type templatedEmailer struct {
	papercloudEmailer mailgun.Emailer
	maplefileEmailer  mailgun.Emailer
}

func NewTemplatedEmailer(papercloudEmailer mailgun.Emailer, maplefileEmailer mailgun.Emailer) TemplatedEmailer {
	return &templatedEmailer{
		papercloudEmailer: papercloudEmailer,
		maplefileEmailer:  maplefileEmailer,
	}
}
