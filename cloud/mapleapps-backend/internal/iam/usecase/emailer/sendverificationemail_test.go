// internal/iam/usecase/emailer/sendverificationemail_test.go
package emailer

import (
	"context"
	"fmt"
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
		name                   string
		monolithModule         int
		user                   *domain.FederatedUser
		setupMock              func()
		expectedError          string
		expectedPanicSubstring string
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
			expectedError:          "",
			expectedPanicSubstring: "",
		},
		{
			name:                   "Validation Error - Nil user",
			monolithModule:         int(constants.MonolithModulePaperCloud),
			user:                   nil,
			setupMock:              func() {},
			expectedError:          "User is missing value",
			expectedPanicSubstring: "",
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
			setupMock:              func() {},
			expectedError:          "First name is required",
			expectedPanicSubstring: "",
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
			setupMock:              func() {},
			expectedError:          "Email is required",
			expectedPanicSubstring: "",
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
			setupMock:              func() {},
			expectedError:          "Email verification code is required",
			expectedPanicSubstring: "",
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
			setupMock:              func() {},
			expectedError:          "Email verification code type is required",
			expectedPanicSubstring: "",
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
			setupMock:              func() {},
			expectedError:          "",                                                  // No error expected, panic is expected
			expectedPanicSubstring: "invalid memory address or nil pointer dereference", // Expect the panic message substring
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
			expectedError:          "assert.AnError general error for testing",
			expectedPanicSubstring: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			// Wrap the code under test in a function literal to recover from panics
			var panicked bool
			var panicValue any // Use 'any' instead of 'interface{}'
			var err error      // Declare err to capture the return value from Execute

			func() {
				defer func() {
					if r := recover(); r != nil {
						panicked = true
						panicValue = r
					}
				}()
				// Call the method under test
				err = useCase.Execute(
					context.Background(),
					tt.monolithModule,
					tt.user,
				)
			}() // Execute the anonymous function immediately

			if panicked {
				// If a panic occurred
				if tt.expectedPanicSubstring != "" {
					// If we expected a panic, check the panic value
					panicMsg := fmt.Sprintf("%v", panicValue)
					assert.Contains(t, panicMsg, tt.expectedPanicSubstring, "Expected panic with substring")
					// If panic was expected and matched, the test passes for this part.
				} else {
					// If we did NOT expect a panic, fail the test
					assert.Fail(t, fmt.Sprintf("Test panicked unexpectedly with value: %v", panicValue))
				}
			} else {
				// If no panic occurred
				if tt.expectedPanicSubstring != "" {
					// If we expected a panic but didn't get one, fail
					assert.Fail(t, "Expected a panic, but method returned without panicking")
				} else if tt.expectedError == "" {
					// If no panic was expected and no error was expected
					assert.NoError(t, err)
				} else {
					// If no panic was expected and an error was expected
					assert.Error(t, err)
					// Check that the error message contains the expected substring
					assert.Contains(t, err.Error(), tt.expectedError)
				}
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
