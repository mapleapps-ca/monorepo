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

func (impl *templatedEmailer) SendUserLoginOneTimeTokenEmail(ctx context.Context, monolithModule int, email, oneTimeToken, firstName string) error {
	switch monolithModule {
	case int(constants.MonolithModuleMapleFile):
		return impl.SendMapleFileModuleUserLoginOneTimeTokenEmail(ctx, email, oneTimeToken, firstName)
	case int(constants.MonolithModulePaperCloud):
		return impl.SendPaperCloudModuleUserLoginOneTimeTokenEmail(ctx, email, oneTimeToken, firstName)
	default:
		return fmt.Errorf("unsupported monolith module: %d", monolithModule)
	}
}

func (impl *templatedEmailer) SendMapleFileModuleUserLoginOneTimeTokenEmail(ctx context.Context, email, oneTimeToken, firstName string) error {
	fp := path.Join("templates", "papercloud/login_ott.html")
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		return fmt.Errorf("user login one-time token parsing error: %w", err)
	}

	var processed bytes.Buffer

	// Render the HTML template with our data.
	data := struct {
		FirstName string
		OTT       string
		Email     string
	}{
		FirstName: firstName,
		OTT:       oneTimeToken,
		Email:     email,
	}
	if err := tmpl.Execute(&processed, data); err != nil {
		return fmt.Errorf("user verification template execution error: %w", err)
	}
	body := processed.String() // DEVELOPERS NOTE: Convert our long sequence of data into a string.

	if err := impl.papercloudEmailer.Send(ctx, impl.papercloudEmailer.GetSenderEmail(), "Login token", email, body); err != nil {
		return fmt.Errorf("sending papercloud login one-time token verification error: %w", err)
	}

	// For debugging purposes only.
	log.Printf("success in sending papercloud login one-time token email: %v\n", oneTimeToken)
	return nil
}

func (impl *templatedEmailer) SendPaperCloudModuleUserLoginOneTimeTokenEmail(ctx context.Context, email, oneTimeToken, firstName string) error {
	fp := path.Join("templates", "maplefile/login_ott.html")
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		return fmt.Errorf("user login one-time token parsing error: %w", err)
	}

	var processed bytes.Buffer

	// Render the HTML template with our data.
	data := struct {
		FirstName string
		OTT       string
		Email     string
	}{
		FirstName: firstName,
		OTT:       oneTimeToken,
		Email:     email,
	}
	if err := tmpl.Execute(&processed, data); err != nil {
		return fmt.Errorf("user verification template execution error: %w", err)
	}
	body := processed.String() // DEVELOPERS NOTE: Convert our long sequence of data into a string.

	if err := impl.maplefileEmailer.Send(ctx, impl.maplefileEmailer.GetSenderEmail(), "Login token", email, body); err != nil {
		return fmt.Errorf("sending maplefile login one-time token verification error: %w", err)
	}
	log.Println("success in sending maplefile login one-time token email")
	return nil
}
