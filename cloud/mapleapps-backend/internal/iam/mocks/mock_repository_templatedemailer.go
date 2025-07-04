// Code generated by MockGen. DO NOT EDIT.
// Source: internal/iam/repo/templatedemailer/interface.go
//
// Generated by this command:
//
//	mockgen -source=internal/iam/repo/templatedemailer/interface.go -destination=internal/iam/mocks/mock_repository_templatedemailer.go -package=mocks
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockTemplatedEmailer is a mock of TemplatedEmailer interface.
type MockTemplatedEmailer struct {
	ctrl     *gomock.Controller
	recorder *MockTemplatedEmailerMockRecorder
	isgomock struct{}
}

// MockTemplatedEmailerMockRecorder is the mock recorder for MockTemplatedEmailer.
type MockTemplatedEmailerMockRecorder struct {
	mock *MockTemplatedEmailer
}

// NewMockTemplatedEmailer creates a new mock instance.
func NewMockTemplatedEmailer(ctrl *gomock.Controller) *MockTemplatedEmailer {
	mock := &MockTemplatedEmailer{ctrl: ctrl}
	mock.recorder = &MockTemplatedEmailerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTemplatedEmailer) EXPECT() *MockTemplatedEmailerMockRecorder {
	return m.recorder
}

// SendUserLoginOneTimeTokenEmail mocks base method.
func (m *MockTemplatedEmailer) SendUserLoginOneTimeTokenEmail(ctx context.Context, monolithModule int, email, oneTimeToken, firstName string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendUserLoginOneTimeTokenEmail", ctx, monolithModule, email, oneTimeToken, firstName)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendUserLoginOneTimeTokenEmail indicates an expected call of SendUserLoginOneTimeTokenEmail.
func (mr *MockTemplatedEmailerMockRecorder) SendUserLoginOneTimeTokenEmail(ctx, monolithModule, email, oneTimeToken, firstName any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendUserLoginOneTimeTokenEmail", reflect.TypeOf((*MockTemplatedEmailer)(nil).SendUserLoginOneTimeTokenEmail), ctx, monolithModule, email, oneTimeToken, firstName)
}

// SendUserPasswordResetEmail mocks base method.
func (m *MockTemplatedEmailer) SendUserPasswordResetEmail(ctx context.Context, monolithModule int, email, verificationCode, firstName string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendUserPasswordResetEmail", ctx, monolithModule, email, verificationCode, firstName)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendUserPasswordResetEmail indicates an expected call of SendUserPasswordResetEmail.
func (mr *MockTemplatedEmailerMockRecorder) SendUserPasswordResetEmail(ctx, monolithModule, email, verificationCode, firstName any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendUserPasswordResetEmail", reflect.TypeOf((*MockTemplatedEmailer)(nil).SendUserPasswordResetEmail), ctx, monolithModule, email, verificationCode, firstName)
}

// SendUserVerificationEmail mocks base method.
func (m *MockTemplatedEmailer) SendUserVerificationEmail(ctx context.Context, monolithModule int, email, verificationCode, firstName string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendUserVerificationEmail", ctx, monolithModule, email, verificationCode, firstName)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendUserVerificationEmail indicates an expected call of SendUserVerificationEmail.
func (mr *MockTemplatedEmailerMockRecorder) SendUserVerificationEmail(ctx, monolithModule, email, verificationCode, firstName any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendUserVerificationEmail", reflect.TypeOf((*MockTemplatedEmailer)(nil).SendUserVerificationEmail), ctx, monolithModule, email, verificationCode, firstName)
}
