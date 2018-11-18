package service

import (
	"errors"
	"testing"
)

func TestNewLineJwtService(t *testing.T) {
	service := NewLineJwtService()
	if service.ClientId == "" || service.ClientSecret == "" {
		t.Errorf("NewLineJwtService()) == %v", nil)
	}
}

func TestLineJwtServiceExtractIdToken(t *testing.T) {
	service := NewLineJwtService()

	// Expire
	token := "eyJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJodHRwczovL2FjY2Vzcy5saW5lLm1lIiwic3ViIjoiVTVmYTliMTUzNDc3OGMyN2QxMDQxNDM2MTRkMTdmYWRkIiwiYXVkIjoiMTYyMjg2NDY4NSIsImV4cCI6MTU0MjUxNDU4OSwiaWF0IjoxNTQyNTEwOTg5LCJuYW1lIjoiQ2hvb3BvbmciLCJwaWN0dXJlIjoiaHR0cHM6Ly9wcm9maWxlLmxpbmUtc2Nkbi5uZXQvMGh4Q1hiMDJ0MEoyeGxHd2o5M3N4WU8xbGVLUUVTTlNFa0hTODhBeFVUY1FsSktUUXlYM282WFJBU2V3bE1lMlJvRG5zNkRrRWZlZ2hQIn0.EVoyJnUj4LhzZPxeXrnC8VN5nnK3sjP9ukno6Rxlf9I"
	wantErr := errors.New("jwt: exp claim is invalid")
	_, err := service.ExtractIdToken(token)
	if err == nil || err.Error() != wantErr.Error() {
		t.Errorf("LineJwtService.ExtractIdToken()) == %v want %v", err, wantErr)
	}

	// Invalid
	token = "invalid token"
	wantErr = errors.New("jwt: malformed token")
	_, err = service.ExtractIdToken(token)
	if err == nil || err.Error() != wantErr.Error() {
		t.Errorf("LineJwtService.ExtractIdToken()) == %v want %v", err, wantErr)
	}
}
