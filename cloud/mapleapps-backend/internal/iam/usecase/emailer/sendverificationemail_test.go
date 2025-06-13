// internal/iam/usecase/emailer/sendverificationemail_test.go
package emailer

import (
	"context"
	"testing"
	"time"

	"github.com/gocql/gocql"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
	domain "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/federateduser"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/mocks"
)

func TestSendFederatedUserVerificationEmailUseCase_Execute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEmailer := mocks.NewMockTemplatedEmailer(ctrl)
	logger := zap.NewNop()
	config := &config.Configuration{}

	useCase := NewSendFederatedUserVerificationEmailUseCase(config, logger, mockEmailer)

	// Create a valid user for testing
	validUser := &domain.FederatedUser{
		ID:        gocql.TimeUUID(),
		Email:     "test@example.com",
		FirstName: "Jane",
		LastName:  "Smith",
		SecurityData: &domain.FederatedUserSecurityData{
			Code:       "verify123",
			CodeType:   domain.FederatedUserCodeTypeEmailVerification,
			CodeExpiry: time.Now().Add(72 * time.Hour),
		},
	}

	tests := []struct {
		name           string
		monolithModule int
		user           *domain.FederatedUser
		setupMock      func()
		expectedError  string
	}{
		{
			name:           "Success - Valid user and email verification",
			monolithModule: int(constants.MonolithModulePaperCloud),
			user:           validUser,
			setupMock: func() {
				mockEmailer.EXPECT().
					SendUserVerificationEmail(
						gomock.Any(),
						int(constants.MonolithModulePaperCloud),
						"test@example.com",
						"verify123",
						"Jane",
					).
					Return(nil).
					Times(1)
			},
			expectedError: "",
		},
		{
			name:           "Validation Error - Nil user",
			monolithModule: int(constants.MonolithModulePaperCloud),
			user:           nil,
			setupMock:      func() {},
			expectedError:  "User is missing value",
		},
		{
			name:           "Validation Error - Missing first name",
			monolithModule: int(constants.MonolithModulePaperCloud),
			user: &domain.FederatedUser{
				ID:        gocql.TimeUUID(),
				Email:     "test@example.com",
				FirstName: "",
				SecurityData: &domain.FederatedUserSecurityData{
					Code:     "verify123",
					CodeType: domain.FederatedUserCodeTypeEmailVerification,
				},
			},
			setupMock:     func() {},
			expectedError: "First name is required",
		},
		{
			name:           "Validation Error - Missing email",
			monolithModule: int(constants.MonolithModulePaperCloud),
			user: &domain.FederatedUser{
				ID:        gocql.TimeUUID(),
				Email:     "",
				FirstName: "Jane",
				SecurityData: &domain.FederatedUserSecurityData{
					Code:     "verify123",
					CodeType: domain.FederatedUserCodeTypeEmailVerification,
				},
			},
			setupMock:     func() {},
			expectedError: "Email is required",
		},
		{
			name:           "Validation Error - Missing verification code",
			monolithModule: int(constants.MonolithModulePaperCloud),
			user: &domain.FederatedUser{
				ID:        gocql.TimeUUID(),
				Email:     "test@example.com",
				FirstName: "Jane",
				SecurityData: &domain.FederatedUserSecurityData{
					Code:     "",
					CodeType: domain.FederatedUserCodeTypeEmailVerification,
				},
			},
			setupMock:     func() {},
			expectedError: "Email verification code is required",
		},
		{
			name:           "Validation Error - Wrong code type",
			monolithModule: int(constants.MonolithModulePaperCloud),
			user: &domain.FederatedUser{
				ID:        gocql.TimeUUID(),
				Email:     "test@example.com",
				FirstName: "Jane",
				SecurityData: &domain.FederatedUserSecurityData{
					Code:     "verify123",
					CodeType: domain.FederatedUserCodeTypePasswordReset, // Wrong type
				},
			},
			setupMock:     func() {},
			expectedError: "Email verification code type is required",
		},
		{
			name:           "Validation Error - Missing security data",
			monolithModule: int(constants.MonolithModulePaperCloud),
			user: &domain.FederatedUser{
				ID:           gocql.TimeUUID(),
				Email:        "test@example.com",
				FirstName:    "Jane",
				SecurityData: nil,
			},
			setupMock:     func() {},
			expectedError: "Email verification data is missing", // Updated expected error
		},
		{
			name:           "Email Service Error",
			monolithModule: int(constants.MonolithModulePaperCloud),
			user:           validUser,
			setupMock: func() {
				mockEmailer.EXPECT().
					SendUserVerificationEmail(
						gomock.Any(),
						int(constants.MonolithModulePaperCloud),
						"test@example.com",
						"verify123",
						"Jane",
					).
					Return(assert.AnError).
					Times(1)
			},
			expectedError: "assert.AnError general error for testing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			err := useCase.Execute(
				context.Background(),
				tt.monolithModule,
				tt.user,
			)

			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				// Check that the error message contains the expected substring
				assert.Contains(t, err.Error(), tt.expectedError)
			}
		})
	}
}

// Test helper functions and benchmarks
func BenchmarkSendLoginOTTEmailUseCase_Execute(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	mockEmailer := mocks.NewMockTemplatedEmailer(ctrl)
	logger := zap.NewNop()
	config := &config.Configuration{}

	useCase := NewSendLoginOTTEmailUseCase(config, logger, mockEmailer)

	// Setup mock to always succeed
	mockEmailer.EXPECT().
		SendUserLoginOneTimeTokenEmail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = useCase.Execute(
				context.Background(),
				int(constants.MonolithModuleMapleFile),
				"test@example.com",
				"123456",
				"John",
			)
		}
	})
}
