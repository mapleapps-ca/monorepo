// internal/iam/usecase/emailer/sendloginott_test.go
package emailer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/mocks"
)

func TestSendLoginOTTEmailUseCase_Execute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEmailer := mocks.NewMockTemplatedEmailer(ctrl)
	logger := zap.NewNop()
	config := &config.Configuration{}

	useCase := NewSendLoginOTTEmailUseCase(config, logger, mockEmailer)

	tests := []struct {
		name           string
		monolithModule int
		email          string
		ott            string
		firstName      string
		setupMock      func()
		expectedError  string
	}{
		{
			name:           "Success - Valid inputs",
			monolithModule: int(constants.MonolithModuleMapleFile),
			email:          "test@example.com",
			ott:            "123456",
			firstName:      "John",
			setupMock: func() {
				mockEmailer.EXPECT().
					SendUserLoginOneTimeTokenEmail(
						gomock.Any(),
						int(constants.MonolithModuleMapleFile),
						"test@example.com",
						"123456",
						"John",
					).
					Return(nil).
					Times(1)
			},
			expectedError: "",
		},
		{
			name:           "Validation Error - Missing first name",
			monolithModule: int(constants.MonolithModuleMapleFile),
			email:          "test@example.com",
			ott:            "123456",
			firstName:      "",
			setupMock:      func() {},
			expectedError:  "First name is required",
		},
		{
			name:           "Validation Error - Missing email",
			monolithModule: int(constants.MonolithModuleMapleFile),
			email:          "",
			ott:            "123456",
			firstName:      "John",
			setupMock:      func() {},
			expectedError:  "Email is required",
		},
		{
			name:           "Validation Error - Missing OTT",
			monolithModule: int(constants.MonolithModuleMapleFile),
			email:          "test@example.com",
			ott:            "",
			firstName:      "John",
			setupMock:      func() {},
			expectedError:  "One-time token is required",
		},
		{
			name:           "Validation Error - Multiple missing fields",
			monolithModule: int(constants.MonolithModuleMapleFile),
			email:          "",
			ott:            "",
			firstName:      "",
			setupMock:      func() {},
			// The actual error for multiple missing fields is a JSON string containing specific errors.
			// Update the expected error to match one of the specific error messages present in the JSON,
			// as assert.Contains checks for substring presence.
			expectedError: "Email is required",
		},
		{
			name:           "Email Service Error",
			monolithModule: int(constants.MonolithModuleMapleFile),
			email:          "test@example.com",
			ott:            "123456",
			firstName:      "John",
			setupMock: func() {
				mockEmailer.EXPECT().
					SendUserLoginOneTimeTokenEmail(
						gomock.Any(),
						int(constants.MonolithModuleMapleFile),
						"test@example.com",
						"123456",
						"John",
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
				tt.email,
				tt.ott,
				tt.firstName,
			)

			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			}
		})
	}
}
