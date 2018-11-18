package service

import (
	"errors"
	"testing"
)

func TestNewLineOAuthService(t *testing.T) {
	service := NewLineOAuthService()
	if service.oAuthConfig == nil {
		t.Errorf("NewLineOAuthService()) == %v", nil)
	}
}

func TestLineOAuthServiceGenerateOAuthState(t *testing.T) {
	service := NewLineOAuthService()
	state := service.GenerateOAuthState()
	if state == "" {
		t.Errorf("LineOAuthService.GenerateOAuthState()) == %v", "")
	}
}

func TestSignout(t *testing.T) {
	service := NewLineOAuthService()
	err := service.Signout("dummy")
	wantErr := errors.New(`{"error":"invalid_request","error_description":"The access token malformed"}`)
	if err == nil || err.Error() != wantErr.Error() {
		t.Errorf("LineOAuthService.Signout()) == %v want %v", err, wantErr)
	}
}
