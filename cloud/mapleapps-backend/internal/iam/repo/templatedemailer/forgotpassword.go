package templatedemailer

import (
	"bytes"
	"context"
	"fmt"
	"path"
	"text/template"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
)

func (impl *templatedEmailer) SendUserPasswordResetEmail(ctx context.Context, monolithModule int, email, verificationCode, firstName string) error {
	switch monolithModule {
	case int(constants.MonolithModuleMapleFile):
		return impl.SendMapleFileUserPasswordResetEmail(ctx, email, verificationCode, firstName)
	case int(constants.MonolithModulePaperCloud):
		return impl.SendPaperCloudUserPasswordResetEmail(ctx, email, verificationCode, firstName)
	default:
		return fmt.Errorf("unsupported monolith module: %d", monolithModule)
	}
}

func (impl *templatedEmailer) SendMapleFileUserPasswordResetEmail(ctx context.Context, email, verificationCode, firstName string) error {

	fp := path.Join("templates", "maplefile/forgot_password.html")
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		return err
	}

	var processed bytes.Buffer

	// Render the HTML template with our data.
	data := struct {
		Email            string
		VerificationCode string
		FirstName        string
	}{
		Email:            email,
		VerificationCode: verificationCode,
		FirstName:        firstName,
	}
	if err := tmpl.Execute(&processed, data); err != nil {
		return err
	}
	body := processed.String() // DEVELOPERS NOTE: Convert our long sequence of data into a string.

	if err := impl.maplefileEmailer.Send(ctx, impl.maplefileEmailer.GetSenderEmail(), "Password Reset", email, body); err != nil {
		return err
	}
	return nil
}

func (impl *templatedEmailer) SendPaperCloudUserPasswordResetEmail(ctx context.Context, email, verificationCode, firstName string) error {

	fp := path.Join("templates", "papercloud/forgot_password.html")
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		return err
	}

	var processed bytes.Buffer

	// Render the HTML template with our data.
	data := struct {
		Email            string
		VerificationCode string
		FirstName        string
	}{
		Email:            email,
		VerificationCode: verificationCode,
		FirstName:        firstName,
	}
	if err := tmpl.Execute(&processed, data); err != nil {
		return err
	}
	body := processed.String() // DEVELOPERS NOTE: Convert our long sequence of data into a string.

	if err := impl.papercloudEmailer.Send(ctx, impl.papercloudEmailer.GetSenderEmail(), "Password Reset", email, body); err != nil {
		return err
	}
	return nil
}
