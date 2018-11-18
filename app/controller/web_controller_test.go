package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/choobot/choo-todo-bot/app/model"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2"
)

type mockTodoModel struct {
	willError       bool
	willNoRemaining bool
}

var todos = []model.Todo{
	{
		ID:     1,
		UserID: "user id",
		Task:   "task",
		Done:   false,
		Pin:    true,
		Due:    time.Now(),
	},
}

func (this *mockTodoModel) List(userID string) ([]model.Todo, error) {
	if this.willError {
		this.willError = false
		return nil, errors.New("dummy")
	}
	return todos, nil
}
func (this *mockTodoModel) Create(todo model.Todo) error {
	if this.willError {
		this.willError = false
		return errors.New("dummy")
	}
	return nil
}
func (this *mockTodoModel) Pin(todo model.Todo) error {
	if this.willError {
		this.willError = false
		return errors.New("dummy")
	}
	return nil
}
func (this *mockTodoModel) Done(todo model.Todo) error {
	if this.willError {
		this.willError = false
		return errors.New("dummy")
	}
	return nil
}
func (this *mockTodoModel) Remind() (map[string][]model.Todo, error) {
	if this.willError {
		this.willError = false
		return nil, errors.New("dummy")
	}
	userTodos := map[string][]model.Todo{}
	todos := userTodos["dummy"]
	if this.willNoRemaining {
		this.willNoRemaining = false
		todo := model.Todo{
			Done: true,
		}
		todos = append(todos, todo)
	} else {
		todo := model.Todo{
			Done: false,
		}
		todos = append(todos, todo)
	}

	todo := model.Todo{
		Pin: true,
	}
	todos = append(todos, todo)
	userTodos["dummy"] = todos
	return userTodos, nil
}

type mockSessionService struct {
	sessions map[string]interface{}
}

func (this *mockSessionService) Get(c echo.Context, name string) interface{} {
	return this.sessions[name]
}

func (this *mockSessionService) Mock(name string, value interface{}) {
	this.sessions[name] = value
}

func (this *mockSessionService) Destroy(c echo.Context) {

}

func (this *mockSessionService) Set(c echo.Context, name string, value interface{}) {
	this.sessions[name] = value
}

type mockOAuthService struct {
	mock.Mock
}

func (this *mockOAuthService) GenerateOAuthState() string {
	return "dummy"
}

func (this *mockOAuthService) OAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     "dummy",
		ClientSecret: "dummy",
		Scopes:       []string{"dummy"},
		RedirectURL:  "https://dummy/redirect",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://dummy/auth",
			TokenURL: "https://dummy/token",
		},
	}
}
func (this *mockOAuthService) Signout(oauthToken string) error {
	args := this.Called(oauthToken)
	return args.Error(0)
}

func TestWebControllerIndex(t *testing.T) {
	todoModel := mockTodoModel{}
	sessionService := mockSessionService{
		sessions: map[string]interface{}{},
	}
	controller := WebController{
		TodoModel:      &todoModel,
		SessionService: &sessionService,
	}
	e := echo.New()

	sessionService.Mock("oauthToken", "dummy")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if assert.NoError(t, controller.Index(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	sessionService.Mock("oauthToken", nil)
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	if assert.NoError(t, controller.Index(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}

func TestWebControllerList(t *testing.T) {
	todoModel := mockTodoModel{}
	sessionService := mockSessionService{
		sessions: map[string]interface{}{},
	}
	controller := WebController{
		TodoModel:      &todoModel,
		SessionService: &sessionService,
	}
	e := echo.New()

	// OK
	sessionService.Mock("oauthId", "dummy")
	b, _ := json.Marshal(todos)
	wantJSON := string(b)
	req := httptest.NewRequest(http.MethodGet, "/list", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, controller.List(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, wantJSON, rec.Body.String())
	}

	// No userID
	sessionService.Mock("oauthId", nil)
	req = httptest.NewRequest(http.MethodGet, "/list", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	if assert.NoError(t, controller.List(c)) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, "user not found", rec.Body.String())
	}

	// Error from Model
	sessionService.Mock("oauthId", "dummy")
	todoModel.willError = true
	req = httptest.NewRequest(http.MethodGet, "/list", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	if assert.NoError(t, controller.List(c)) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, "dummy", rec.Body.String())
	}
}

func TestWebControllerPin(t *testing.T) {
	todoModel := mockTodoModel{}
	sessionService := mockSessionService{
		sessions: map[string]interface{}{},
	}
	controller := WebController{
		TodoModel:      &todoModel,
		SessionService: &sessionService,
	}
	e := echo.New()

	// Valid
	b, _ := json.Marshal(todos[0])
	inputJSON := string(b)
	req := httptest.NewRequest(http.MethodPost, "/pin", strings.NewReader(inputJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, controller.Pin(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "", rec.Body.String())
	}

	// Invalid JSON
	b, _ = json.Marshal("")
	inputJSON = string(b)
	req = httptest.NewRequest(http.MethodPost, "/pin", strings.NewReader(inputJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	if assert.NoError(t, controller.Pin(c)) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, "code=400, message=Unmarshal type error: expected=model.Todo, got=string, field=, offset=2", rec.Body.String())
	}

	// Error from Model
	todoModel.willError = true
	b, _ = json.Marshal(todos[0])
	inputJSON = string(b)
	req = httptest.NewRequest(http.MethodPost, "/pin", strings.NewReader(inputJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	if assert.NoError(t, controller.Pin(c)) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, "dummy", rec.Body.String())
	}
}

func TestWebControllerDone(t *testing.T) {
	todoModel := mockTodoModel{}
	sessionService := mockSessionService{
		sessions: map[string]interface{}{},
	}
	controller := WebController{
		TodoModel:      &todoModel,
		SessionService: &sessionService,
	}
	e := echo.New()

	// Valid
	b, _ := json.Marshal(todos[0])
	inputJSON := string(b)
	req := httptest.NewRequest(http.MethodPost, "/pin", strings.NewReader(inputJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, controller.Done(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "", rec.Body.String())
	}

	// Invalid JSON
	b, _ = json.Marshal("")
	inputJSON = string(b)
	req = httptest.NewRequest(http.MethodPost, "/done", strings.NewReader(inputJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	if assert.NoError(t, controller.Done(c)) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, "code=400, message=Unmarshal type error: expected=model.Todo, got=string, field=, offset=2", rec.Body.String())
	}

	// Error from Model
	todoModel.willError = true
	b, _ = json.Marshal(todos[0])
	inputJSON = string(b)
	req = httptest.NewRequest(http.MethodPost, "/pin", strings.NewReader(inputJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	if assert.NoError(t, controller.Done(c)) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, "dummy", rec.Body.String())
	}
}

func TestWebControllerUserInfo(t *testing.T) {
	todoModel := mockTodoModel{}
	sessionService := mockSessionService{
		sessions: map[string]interface{}{},
	}
	controller := WebController{
		TodoModel:      &todoModel,
		SessionService: &sessionService,
	}
	e := echo.New()

	// OK
	sessionService.Mock("oauthName", "dummy")
	sessionService.Mock("oauthPicture", "dummy")
	userInfo := map[string]string{
		"oauthName":    "dummy",
		"oauthPicture": "dummy",
	}
	b, _ := json.Marshal(userInfo)
	wantJSON := string(b)
	req := httptest.NewRequest(http.MethodGet, "/list", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, controller.UserInfo(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, wantJSON, rec.Body.String())
	}

	// No user
	sessionService.Mock("oauthName", nil)
	sessionService.Mock("oauthPicture", nil)
	req = httptest.NewRequest(http.MethodGet, "/list", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	if assert.NoError(t, controller.UserInfo(c)) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, "user not found", rec.Body.String())
	}
}

func TestWebControllerLogout(t *testing.T) {
	todoModel := mockTodoModel{}
	sessionService := mockSessionService{
		sessions: map[string]interface{}{},
	}

	e := echo.New()

	// Valid
	oAuthService := mockOAuthService{}
	oAuthService.On("Signout", mock.Anything).Return(nil)

	controller := WebController{
		TodoModel:      &todoModel,
		SessionService: &sessionService,
		OAuthService:   &oAuthService,
	}

	sessionService.Mock("oauthToken", "dummy")
	req := httptest.NewRequest(http.MethodGet, "/logout", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, controller.Logout(c)) {
		assert.Equal(t, http.StatusTemporaryRedirect, rec.Code)
		assert.Equal(t, "/", rec.Header().Get("Location"))
	}

	// Invalid
	oAuthService = mockOAuthService{}
	oAuthService.On("Signout", mock.Anything).Return(errors.New("dummy"))
	controller = WebController{
		TodoModel:      &todoModel,
		SessionService: &sessionService,
		OAuthService:   &oAuthService,
	}
	sessionService.Mock("oauthToken", "dummy")
	req = httptest.NewRequest(http.MethodGet, "/logout", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	if assert.NoError(t, controller.Logout(c)) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	}
}

func TestWebControllerLogin(t *testing.T) {
	todoModel := mockTodoModel{}
	sessionService := mockSessionService{
		sessions: map[string]interface{}{},
	}

	e := echo.New()
	// Valid
	oAuthService := mockOAuthService{}

	controller := WebController{
		TodoModel:      &todoModel,
		SessionService: &sessionService,
		OAuthService:   &oAuthService,
	}

	sessionService.Mock("oauthToken", "dummy")
	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, controller.Login(c)) {
		assert.Equal(t, http.StatusTemporaryRedirect, rec.Code)
		assert.Equal(t, "https://dummy/auth?access_type=online&client_id=dummy&redirect_uri=https%3A%2F%2Fdummy%2Fredirect&response_type=code&scope=dummy&state=dummy", rec.Header().Get("Location"))
	}
}

func TestWebControllerAuth(t *testing.T) {
	todoModel := mockTodoModel{}
	sessionService := mockSessionService{
		sessions: map[string]interface{}{},
	}

	e := echo.New()
	// Valid
	oAuthService := mockOAuthService{}
	oAuthService.On("Signout", mock.Anything).Return(nil)

	controller := WebController{
		TodoModel:      &todoModel,
		SessionService: &sessionService,
		OAuthService:   &oAuthService,
	}

	// Error from OAuth
	sessionService.Mock("oauthState", "dummy")
	req := httptest.NewRequest(http.MethodGet, "/auth?state=dummy", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, controller.Auth(c)) {
		// assert.Equal(t, http.StatusOK, rec.Code)
	}

	// No oauthState
	sessionService.Mock("oauthState", nil)
	req = httptest.NewRequest(http.MethodGet, "/auth?state=dummy", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	if assert.NoError(t, controller.Auth(c)) {
		assert.Equal(t, http.StatusTemporaryRedirect, rec.Code)
		assert.Equal(t, "/", rec.Header().Get("Location"))
	}
}
