package service

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

type OAuthService interface {
	GenerateOAuthState() string
	OAuthConfig() *oauth2.Config
	Signout(oauthToken string) error
}

type LineOAuthService struct {
	oAuthConfig *oauth2.Config
}

func (this *LineOAuthService) GenerateOAuthState() string {
	salt := "choo-todo-bot"
	data := []byte(strconv.Itoa(int(time.Now().Unix())) + salt)
	return fmt.Sprintf("%x", sha1.Sum(data))
}

func (this *LineOAuthService) OAuthConfig() *oauth2.Config {
	return this.oAuthConfig
}

func (this *LineOAuthService) Signout(oauthToken string) error {
	form := url.Values{}
	form.Add("access_token", oauthToken)
	form.Add("client_id", this.OAuthConfig().ClientID)
	form.Add("client_secret", this.OAuthConfig().ClientSecret)
	req, err := http.NewRequest("POST", "https://api.line.me/oauth2/v2.1/revoke", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		defer res.Body.Close()
		body, _ := ioutil.ReadAll(res.Body)
		return errors.New((string(body)))
	}
	return nil
}

func NewLineOAuthService() LineOAuthService {
	oAuthConfig := oauth2.Config{
		ClientID:     os.Getenv("LINE_LOGIN_ID"),
		ClientSecret: os.Getenv("LINE_LOGIN_SECRET"),
		Scopes:       []string{"openid", "profile"},
		RedirectURL:  os.Getenv("LINE_LOGIN_REDIRECT_URL"),
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://access.line.me/oauth2/v2.1/authorize",
			TokenURL: "https://api.line.me/oauth2/v2.1/token",
		},
	}
	return LineOAuthService{
		oAuthConfig: &oAuthConfig,
	}
}
