package templatedemailer

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"path"
	"text/template"
)

func (impl *templatedEmailer) SendUserLoginOneTimeTokenEmail(ctx context.Context, monolithModule int, email, oneTimeToken, firstName string) error {
	switch monolithModule {
	case 1:
		return impl.SendPaperCloudPropertyEvaluatorModuleUserLoginOneTimeTokenEmail(ctx, email, oneTimeToken, firstName)
	default:
		return fmt.Errorf("unsupported monolith module: %d", monolithModule)
	}
}

func (impl *templatedEmailer) SendPaperCloudPropertyEvaluatorModuleUserLoginOneTimeTokenEmail(ctx context.Context, email, oneTimeToken, firstName string) error {
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

	if err := impl.incomePropertyEmailer.Send(ctx, impl.incomePropertyEmailer.GetSenderEmail(), "Login token", email, body); err != nil {
		return fmt.Errorf("sending income property evaluator login one-time token verification error: %w", err)
	}
	log.Println("success in sending income property evaluator login one-time token email")
	return nil
}
