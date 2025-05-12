package templatedemailer

import (
	"context"

	"go.uber.org/zap"
)

// TemplatedEmailer Is adapter for responsive HTML email templates sender.
type TemplatedEmailer interface {
	GetBackendDomainName() string
	GetFrontendDomainName() string
	// SendBusinessVerificationEmail(email, verificationCode, firstName string) error
	SendUserVerificationEmail(ctx context.Context, email, verificationCode, firstName string) error
	// SendNewUserTemporaryPasswordEmail(email, firstName, temporaryPassword string) error
	SendUserPasswordResetEmail(ctx context.Context, email, verificationCode, firstName string) error
	// SendNewComicSubmissionEmailToStaff(staffEmails []string, submissionID string, storeName string, item string, cpsrn string, serviceTypeName string) error
	// SendNewComicSubmissionEmailToRetailers(retailerEmails []string, submissionID string, storeName string, item string, cpsrn string, serviceTypeName string) error
	// SendNewStoreEmailToStaff(staffEmails []string, storeID string) error
	// SendRetailerStoreActiveEmailToRetailers(retailerEmails []string, storeName string) error
}

type templatedEmailer struct {
	Logger *zap.Logger
}

func NewTemplatedEmailer(logger *zap.Logger) TemplatedEmailer {

	return &templatedEmailer{
		Logger: logger,
	}
}

func (impl *templatedEmailer) GetBackendDomainName() string {
	return ""
}

func (impl *templatedEmailer) GetFrontendDomainName() string {
	return ""
}
