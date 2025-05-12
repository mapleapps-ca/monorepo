package httperror

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		errors   map[string]string
		wantCode int
	}{
		{
			name:     "basic error",
			code:     http.StatusBadRequest,
			errors:   map[string]string{"field": "error message"},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "empty errors map",
			code:     http.StatusNotFound,
			errors:   map[string]string{},
			wantCode: http.StatusNotFound,
		},
		{
			name:     "multiple errors",
			code:     http.StatusBadRequest,
			errors:   map[string]string{"field1": "error1", "field2": "error2"},
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := New(tt.code, &tt.errors)

			httpErr, ok := err.(HTTPError)
			if !ok {
				t.Fatal("expected HTTPError type")
			}
			if httpErr.Code != tt.wantCode {
				t.Errorf("Code = %v, want %v", httpErr.Code, tt.wantCode)
			}
			for k, v := range tt.errors {
				if (*httpErr.Errors)[k] != v {
					t.Errorf("Errors[%s] = %v, want %v", k, (*httpErr.Errors)[k], v)
				}
			}
		})
	}
}

func TestNewForBadRequest(t *testing.T) {
	tests := []struct {
		name   string
		errors map[string]string
	}{
		{
			name:   "single error",
			errors: map[string]string{"field": "error"},
		},
		{
			name:   "multiple errors",
			errors: map[string]string{"field1": "error1", "field2": "error2"},
		},
		{
			name:   "empty errors",
			errors: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewForBadRequest(&tt.errors)

			httpErr, ok := err.(HTTPError)
			if !ok {
				t.Fatal("expected HTTPError type")
			}
			if httpErr.Code != http.StatusBadRequest {
				t.Errorf("Code = %v, want %v", httpErr.Code, http.StatusBadRequest)
			}
			for k, v := range tt.errors {
				if (*httpErr.Errors)[k] != v {
					t.Errorf("Errors[%s] = %v, want %v", k, (*httpErr.Errors)[k], v)
				}
			}
		})
	}
}

func TestNewForSingleField(t *testing.T) {
	tests := []struct {
		name    string
		code    int
		field   string
		message string
	}{
		{
			name:    "basic error",
			code:    http.StatusBadRequest,
			field:   "test",
			message: "error",
		},
		{
			name:    "empty field",
			code:    http.StatusNotFound,
			field:   "",
			message: "error",
		},
		{
			name:    "empty message",
			code:    http.StatusBadRequest,
			field:   "field",
			message: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewForSingleField(tt.code, tt.field, tt.message)

			httpErr, ok := err.(HTTPError)
			if !ok {
				t.Fatal("expected HTTPError type")
			}
			if httpErr.Code != tt.code {
				t.Errorf("Code = %v, want %v", httpErr.Code, tt.code)
			}
			if (*httpErr.Errors)[tt.field] != tt.message {
				t.Errorf("Errors[%s] = %v, want %v", tt.field, (*httpErr.Errors)[tt.field], tt.message)
			}
		})
	}
}

func TestError(t *testing.T) {
	tests := []struct {
		name    string
		errors  map[string]string
		wantErr bool
	}{
		{
			name:    "valid json",
			errors:  map[string]string{"field": "error"},
			wantErr: false,
		},
		{
			name:    "empty map",
			errors:  map[string]string{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := HTTPError{
				Code:   http.StatusBadRequest,
				Errors: &tt.errors,
			}

			errStr := err.Error()
			var jsonMap map[string]string
			if jsonErr := json.Unmarshal([]byte(errStr), &jsonMap); (jsonErr != nil) != tt.wantErr {
				t.Errorf("Error() json.Unmarshal error = %v, wantErr %v", jsonErr, tt.wantErr)
				return
			}

			if !tt.wantErr {
				for k, v := range tt.errors {
					if jsonMap[k] != v {
						t.Errorf("Error() jsonMap[%s] = %v, want %v", k, jsonMap[k], v)
					}
				}
			}
		})
	}
}

func TestResponseError(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantCode    int
		wantContent string
	}{
		{
			name:        "http error",
			err:         NewForBadRequestWithSingleField("field", "invalid"),
			wantCode:    http.StatusBadRequest,
			wantContent: `{"field":"invalid"}`,
		},
		{
			name:        "standard error",
			err:         fmt.Errorf("standard error"),
			wantCode:    http.StatusInternalServerError,
			wantContent: `"standard error"`,
		},
		{
			name:        "nil error",
			err:         errors.New("<nil>"),
			wantCode:    http.StatusInternalServerError,
			wantContent: `"\u003cnil\u003e"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			ResponseError(rr, tt.err)

			// Check status code
			if rr.Code != tt.wantCode {
				t.Errorf("ResponseError() code = %v, want %v", rr.Code, tt.wantCode)
			}

			// Check content type
			if ct := rr.Header().Get("Content-Type"); ct != "Application/json" {
				t.Errorf("ResponseError() Content-Type = %v, want Application/json", ct)
			}

			// Trim newline from response for comparison
			got := rr.Body.String()
			got = got[:len(got)-1] // Remove trailing newline added by json.Encoder
			if got != tt.wantContent {
				t.Errorf("ResponseError() content = %v, want %v", got, tt.wantContent)
			}
		})
	}
}

func TestErrorWrapping(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := fmt.Errorf("wrapped: %w", originalErr)
	httpErr := NewForBadRequestWithSingleField("field", wrappedErr.Error())

	// Test error unwrapping
	if !errors.Is(httpErr, httpErr) {
		t.Error("errors.Is failed for same error")
	}

	var targetErr HTTPError
	if !errors.As(httpErr, &targetErr) {
		t.Error("errors.As failed to get HTTPError")
	}
}

// Test all convenience constructors
func TestConvenienceConstructors(t *testing.T) {
	tests := []struct {
		name     string
		create   func() error
		wantCode int
	}{
		{
			name: "NewForBadRequestWithSingleField",
			create: func() error {
				return NewForBadRequestWithSingleField("field", "message")
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "NewForNotFoundWithSingleField",
			create: func() error {
				return NewForNotFoundWithSingleField("field", "message")
			},
			wantCode: http.StatusNotFound,
		},
		{
			name: "NewForServiceUnavailableWithSingleField",
			create: func() error {
				return NewForServiceUnavailableWithSingleField("field", "message")
			},
			wantCode: http.StatusServiceUnavailable,
		},
		{
			name: "NewForLockedWithSingleField",
			create: func() error {
				return NewForLockedWithSingleField("field", "message")
			},
			wantCode: http.StatusLocked,
		},
		{
			name: "NewForForbiddenWithSingleField",
			create: func() error {
				return NewForForbiddenWithSingleField("field", "message")
			},
			wantCode: http.StatusForbidden,
		},
		{
			name: "NewForUnauthorizedWithSingleField",
			create: func() error {
				return NewForUnauthorizedWithSingleField("field", "message")
			},
			wantCode: http.StatusUnauthorized,
		},
		{
			name: "NewForGoneWithSingleField",
			create: func() error {
				return NewForGoneWithSingleField("field", "message")
			},
			wantCode: http.StatusGone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.create()
			httpErr, ok := err.(HTTPError)
			if !ok {
				t.Fatal("expected HTTPError type")
			}
			if httpErr.Code != tt.wantCode {
				t.Errorf("Code = %v, want %v", httpErr.Code, tt.wantCode)
			}
			if (*httpErr.Errors)["field"] != "message" {
				t.Errorf("Error message = %v, want 'message'", (*httpErr.Errors)["field"])
			}
		})
	}
}
