package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gbrlsnchs/jwt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/middleware"
	"github.com/line/line-bot-sdk-go/linebot"
	"golang.org/x/oauth2"
)

type IdToken struct {
	*jwt.JWT
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

type TodoBot struct {
	client    *linebot.Client
	todoModel TodoModel
}

type Todo struct {
	ID     int
	UserID string
	Task   string
	Done   bool
	Pin    bool
	Due    time.Time
}

type TodoModel interface {
	list(userID string) ([]Todo, error)
	Create(todo Todo) error
	pin(id int) error
	done(id int) error
	remind() (map[string][]Todo, error)
}

type TodoMySqlModel struct {
	db *sql.DB
}

func (this *TodoMySqlModel) list(userID string) ([]Todo, error) {
	return nil, nil
}

func NewTodoMySqlModel() TodoMySqlModel {
	db, _ := sql.Open("mysql", os.Getenv("DATA_SOURCE_NAME"))
	return TodoMySqlModel{
		db: db,
	}
}

func (this *TodoMySqlModel) CreateTablesIfNotExist() error {
	sql := "SELECT 1 FROM todo LIMIT 1"
	_, err := this.db.Query(sql)
	if err != nil {
		sql = `
		CREATE TABLE todo (
			id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			user_id VARCHAR(255) NOT NULL,
			task TEXT NOT NULL,
			done BOOL NOT NULL DEFAULT FALSE,
			pin BOOL NOT NULL DEFAULT FALSE,
			due DATETIME NOT NULL
		) CHARACTER SET utf8 COLLATE utf8_general_ci`

		_, err = this.db.Exec(sql)
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *TodoMySqlModel) Create(todo Todo) error {
	err := this.CreateTablesIfNotExist()
	if err != nil {
		return err
	}

	sql := `INSERT INTO todo ( user_id, task, due ) VALUES( ?, ?, ?)`
	_, err = this.db.Exec(sql, todo.UserID, todo.Task, todo.Due)
	if err != nil {
		return err
	}

	return nil
}

func (this *TodoMySqlModel) pin(id int) error {
	return nil
}

func (this *TodoMySqlModel) done(id int) error {
	return nil
}

func (this *TodoMySqlModel) remind() (map[string][]Todo, error) {
	return nil, nil
}

type OAuthService interface {
	GenerateOAuthState() string
	OAuthConfig() *oauth2.Config
}

type LineOAuthService struct {
	oAuthConfig *oauth2.Config
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

func (this *LineOAuthService) GenerateOAuthState() string {
	//TODO
	return "thisshouldberandom"
}

func (this *LineOAuthService) OAuthConfig() *oauth2.Config {
	return this.oAuthConfig
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

type WebController struct {
	oAuthService OAuthService
	jwtService   JwtService
}

func (this *WebController) SetNoCache(c echo.Context) {
	c.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Response().Header().Set("Pragma", "no-cache")
	c.Response().Header().Set("Expires", "0")
}

func (this *WebController) Index(c echo.Context) error {
	this.SetNoCache(c)
	sess, _ := session.Get("session", c)
	oauthToken := sess.Values["oauthToken"]
	if oauthToken == nil {
		return c.File("views/login.html")
	}
	// return c.File("views/list.html")
	return c.HTML(http.StatusOK, fmt.Sprintf("%#v", sess))
}

func (this *WebController) Login(c echo.Context) error {
	this.SetNoCache(c)
	sess, _ := session.Get("session", c)
	oauthState := this.oAuthService.GenerateOAuthState()
	sess.Values["oauthState"] = oauthState
	url := this.oAuthService.OAuthConfig().AuthCodeURL(oauthState, oauth2.AccessTypeOnline)
	sess.Save(c.Request(), c.Response())
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

func (this *WebController) Auth(c echo.Context) error {
	this.SetNoCache(c)
	sess, _ := session.Get("session", c)
	oauthState := sess.Values["oauthState"]
	state := c.QueryParam("state")
	if oauthState != "" && state != oauthState {
		log.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthState, state)
		return c.Redirect(http.StatusTemporaryRedirect, "/")
	}
	code := c.QueryParam("code")
	oauthToken, err := this.oAuthService.OAuthConfig().Exchange(oauth2.NoContext, code)
	if err != nil {
		return c.HTML(http.StatusInternalServerError, err.Error())
	}

	idToken, err := this.jwtService.ExtractIdToken(oauthToken.Extra("id_token").(string))
	if err != nil {
		return c.HTML(http.StatusInternalServerError, err.Error())
	}

	sess.Values["oauthToken"] = oauthToken.AccessToken
	sess.Values["oauthId"] = idToken.JWT.Subject
	sess.Values["oauthName"] = idToken.Name
	sess.Values["oauthPicture"] = idToken.Picture
	sess.Save(c.Request(), c.Response())
	return c.Redirect(http.StatusTemporaryRedirect, "/")
}

func main() {
	client, err := linebot.New(os.Getenv("LINE_BOT_SECRET"), os.Getenv("LINE_BOT_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	todoModel := NewTodoMySqlModel()

	bot := &TodoBot{
		todoModel: &todoModel,
		client:    client,
	}
	oAuthSerivce := NewLineOAuthService()
	jwtService := NewLineJwtService()
	webController := WebController{
		oAuthService: &oAuthSerivce,
		jwtService:   &jwtService,
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("choo-todo-bot"))))

	// Routes
	e.Static("/", "assets")
	e.GET("/", webController.Index)

	e.POST("/callback", bot.Response)
	e.GET("/remind", bot.Remind)
	e.GET("/login", webController.Login)
	e.GET("/auth", webController.Auth)
	// e.GET("/list", webController.List)
	// e.POST("/pin", webController.Pin)
	// e.POST("/done", webController.Done)
	// e.GET("/user", webController.UserInfo)
	// e.GET("/logout", webController.Logout)

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}
	fmt.Println("Runnign at :" + port)
	e.Logger.Fatal(e.Start(":" + port))
}

func (this *TodoBot) Remind(c echo.Context) error {
	if _, err := this.client.PushMessage("U5fa9b1534778c27d104143614d17fadd", linebot.NewTextMessage("this is reminder")).Do(); err != nil {
		log.Println(err)
	}
	return c.NoContent(http.StatusOK)
}

// 1) Go shopping : 2/5/18 : 13:00
// 2) Go shopping : 2/5/18
// 3) Go shopping : today : 15:30
// 4) Go shopping : today
// 5) Go shopping : tomorrow : 18:00
// 6) Go shopping : tomorrow
func (this *TodoBot) ParseUserMessage(msg string) (Todo, error) {
	loc, _ := time.LoadLocation("Asia/Bangkok")
	getDate := func(word string) string {
		format := "2/1/06"
		if strings.ToLower(word) == "today" {
			return time.Now().In(loc).Format(format)
		} else if strings.ToLower(word) == "tomorrow" {
			return time.Now().In(loc).AddDate(0, 0, 1).Format(format)
		}
		return word
	}
	layout := "2/1/06 15:04"
	words := strings.Split(msg, " : ")
	task := ""
	var due time.Time
	var err error
	if len(words) == 2 {
		task = words[0]
		due, err = time.ParseInLocation(layout, getDate(words[1])+" 12:00", loc)
		if err != nil {
			return Todo{}, errors.New("Wrong format")
		}
	} else if len(words) == 3 {
		task = words[0]
		due, err = time.ParseInLocation(layout, getDate(words[1])+" "+words[2], loc)
		if err != nil {
			return Todo{}, errors.New("Wrong format")
		}
	} else {
		return Todo{}, errors.New("Wrong format")
	}
	todo := Todo{
		Task: task,
		Due:  due,
	}
	return todo, nil
}

func (this *TodoBot) Response(c echo.Context) error {
	howto := `You can create todo list by using these formats:
	1) Go shopping : 2/5/18 : 13:00
	2) Go shopping : 2/5/18
	3) Go shopping : today : 15:30
	4) Go shopping : today
	5) Go shopping : tomorrow : 18:00
	6) Go shopping : tomorrow
You can edit todo list by input word "edit"`
	events, err := this.client.ParseRequest(c.Request())
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			return c.NoContent(http.StatusBadRequest)
		} else {
			return c.NoContent(http.StatusInternalServerError)
		}

	}
	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				msg := message.Text
				if strings.ToLower(msg) == "edit" {
					reply := "Please go to " + os.Getenv("EDIT_URL")
					if _, err = this.client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(reply)).Do(); err != nil {
						log.Println(err)
					}
					return c.NoContent(http.StatusOK)
				} else {
					todo, err := this.ParseUserMessage(msg)
					if err != nil {
						if _, err = this.client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(howto)).Do(); err != nil {
							log.Println(err)
						}
						return c.NoContent(http.StatusOK)
					} else {
						todo.UserID = event.Source.UserID
						if err := this.todoModel.Create(todo); err != nil {
							reply := err.Error()
							if _, err = this.client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(reply)).Do(); err != nil {
								log.Println(err)
							}
							return c.NoContent(http.StatusOK)
						} else {
							reply := "Task has been created."
							if _, err = this.client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(reply)).Do(); err != nil {
								log.Println(err)
							}
							return c.NoContent(http.StatusOK)
						}
					}
				}

			}
		} else if event.Type == linebot.EventTypeJoin {
			replyMessage := "Thanks for adding me. I'm Choo Todo Bot, I'm here to help you to manage your tasks.\n" + howto
			if _, err = this.client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyMessage)).Do(); err != nil {
				log.Println(err)
			}
		}
	}
	return c.NoContent(http.StatusOK)
}
