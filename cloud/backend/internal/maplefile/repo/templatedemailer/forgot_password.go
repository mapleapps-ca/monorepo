package templatedemailer

import (
	"context"
)

func (impl *templatedEmailer) SendUserPasswordResetEmail(ctx context.Context, email, verificationCode, firstName string) error {

	return nil
}
