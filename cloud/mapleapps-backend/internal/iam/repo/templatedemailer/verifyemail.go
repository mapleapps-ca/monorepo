package templatedemailer

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"path"
	"text/template"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
)

func (impl *templatedEmailer) SendUserVerificationEmail(ctx context.Context, monolithModule int, email, verificationCode, firstName string) error {
	switch monolithModule {
	case int(constants.MonolithModuleMapleFile):
		return impl.SendMapleFileModuleUserVerificationEmail(ctx, email, verificationCode, firstName)
	case int(constants.MonolithModulePaperCloud):
		return impl.SendPaperCloudModuleUserVerificationEmail(ctx, email, verificationCode, firstName)
	default:
		return fmt.Errorf("unsupported monolith module: %d", monolithModule)
	}
}

func (impl *templatedEmailer) SendMapleFileModuleUserVerificationEmail(ctx context.Context, email, verificationCode, firstName string) error {
	fp := path.Join("templates", "maplefile/user_verification_email.html")
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		return fmt.Errorf("sending maplefile user verification parsing error: %w", err)
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
		return fmt.Errorf("sending maplefile user verification template execution error: %w", err)
	}
	body := processed.String() // DEVELOPERS NOTE: Convert our long sequence of data into a string.

	if err := impl.maplefileEmailer.Send(ctx, impl.maplefileEmailer.GetSenderEmail(), "Activate your MapleFile Account", email, body); err != nil {
		return fmt.Errorf("sending maplefile user verification error: %w", err)
	}
	log.Printf("success in sending maplefile user verification email: %v\n", verificationCode)
	return nil
}

func (impl *templatedEmailer) SendPaperCloudModuleUserVerificationEmail(ctx context.Context, email, verificationCode, firstName string) error {
	fp := path.Join("templates", "papercloud/user_verification_email.html")
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		return fmt.Errorf("sending papercloud user verification parsing error: %w", err)
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
		return fmt.Errorf("sending papercloud user verification template execution error: %w", err)
	}
	body := processed.String() // DEVELOPERS NOTE: Convert our long sequence of data into a string.

	if err := impl.papercloudEmailer.Send(ctx, impl.papercloudEmailer.GetSenderEmail(), "Activate your PaperCloud Account", email, body); err != nil {
		return fmt.Errorf("sending papercloud user verification error: %w", err)
	}
	log.Printf("success in sending papercloud user verification email: %v\n", verificationCode)
	return nil
}
