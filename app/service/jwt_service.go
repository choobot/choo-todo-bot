package service

import (
	"os"
	"time"

	"github.com/gbrlsnchs/jwt"
)

type IdToken struct {
	*jwt.JWT
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

type JwtService interface {
	ExtractIdToken(tokenValue string) (IdToken, error)
}

func NewLineJwtService() LineJwtService {
	return LineJwtService{
		ClientId:     os.Getenv("LINE_LOGIN_ID"),
		ClientSecret: os.Getenv("LINE_LOGIN_SECRET"),
	}
}

type LineJwtService struct {
	ClientId     string
	ClientSecret string
}

func (this *LineJwtService) ExtractIdToken(tokenValue string) (IdToken, error) {
	var idToken IdToken
	now := time.Now()
	hs256 := jwt.NewHS256(this.ClientSecret)
	payload, sig, err := jwt.Parse(tokenValue)
	if err != nil {
		return idToken, err
	}
	if err = hs256.Verify(payload, sig); err != nil {
		return idToken, err
	}

	if err = jwt.Unmarshal(payload, &idToken); err != nil {
		return idToken, err
	}
	iatValidator := jwt.IssuedAtValidator(now)
	expValidator := jwt.ExpirationTimeValidator(now)
	audValidator := jwt.AudienceValidator(this.ClientId)
	if err = idToken.Validate(iatValidator, expValidator, audValidator); err != nil {
		switch err {
		case jwt.ErrIatValidation:
			return idToken, err
		case jwt.ErrExpValidation:
			return idToken, err
		case jwt.ErrAudValidation:
			return idToken, err
		}
	}
	return idToken, nil
}
