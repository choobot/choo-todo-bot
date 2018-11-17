package main

import (
	"crypto/sha1"
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
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
	List(userID string) ([]Todo, error)
	Create(todo Todo) error
	Pin(todo Todo) error
	Done(todo Todo) error
	Remind() (map[string][]Todo, error)
}

type TodoMySqlModel struct {
	db *sql.DB
}

func (this *TodoMySqlModel) SetTimeZone() error {
	sql := `SET time_zone = 'Asia/Bangkok'`
	_, err := this.db.Exec(sql)
	if err != nil {
		return err
	}

	return nil
}

func (this *TodoMySqlModel) List(userID string) ([]Todo, error) {
	this.SetTimeZone()
	err := this.CreateTablesIfNotExist()
	if err != nil {
		return nil, err
	}
	var todos []Todo
	rows, err := this.db.Query("SELECT id, task, done, pin, due FROM todo WHERE user_id=?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var task string
		var done bool
		var pin bool
		var due time.Time
		if err := rows.Scan(&id, &task, &done, &pin, &due); err != nil {
			return nil, err
		}
		loc, _ := time.LoadLocation("Asia/Bangkok")
		due = due.In(loc)
		todo := Todo{
			ID:     id,
			UserID: userID,
			Task:   task,
			Pin:    pin,
			Done:   done,
			Due:    due,
		}
		todos = append(todos, todo)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return todos, nil
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
	this.SetTimeZone()
	err := this.CreateTablesIfNotExist()
	if err != nil {
		return err
	}
	sql := `INSERT INTO todo ( user_id, task, due ) VALUES( ?, ?, ?)`
	result, err := this.db.Exec(sql, todo.UserID, todo.Task, todo.Due)
	if err != nil {
		return err
	}
	num, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if num != 1 {
		return errors.New("No record")
	}

	return nil
}
func (this *TodoMySqlModel) Pin(todo Todo) error {
	sql := `UPDATE todo SET pin=? WHERE id=?`
	result, err := this.db.Exec(sql, todo.Pin, todo.ID)
	if err != nil {
		return err
	}
	num, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if num != 1 {
		return errors.New("No record")
	}

	return nil
}

func (this *TodoMySqlModel) Done(todo Todo) error {
	sql := `UPDATE todo SET done=? WHERE id=?`
	result, err := this.db.Exec(sql, todo.Done, todo.ID)
	if err != nil {
		return err
	}
	num, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if num != 1 {
		return errors.New("No record")
	}

	return nil
}

func (this *TodoMySqlModel) Remind() (map[string][]Todo, error) {
	this.SetTimeZone()
	userTodos := map[string][]Todo{}
	rows, err := this.db.Query("SELECT user_id, id, task, done, pin, due FROM todo ORDER BY user_id, done, pin DESC, due")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var userID string
		var id int
		var task string
		var done bool
		var pin bool
		var due time.Time
		if err := rows.Scan(&userID, &id, &task, &done, &pin, &due); err != nil {
			return nil, err
		}
		loc, _ := time.LoadLocation("Asia/Bangkok")
		due = due.In(loc)
		todo := Todo{
			ID:     id,
			UserID: userID,
			Task:   task,
			Pin:    pin,
			Done:   done,
			Due:    due,
		}
		log.Println(due)
		//Add to map
		todos := userTodos[userID]
		todos = append(todos, todo)
		userTodos[userID] = todos
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return userTodos, nil
}

type OAuthService interface {
	GenerateOAuthState() string
	OAuthConfig() *oauth2.Config
	Signout(oauthToken string) error
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
		log.Println(string(body))
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

type WebController struct {
	oAuthService OAuthService
	jwtService   JwtService
	todoModel    TodoModel
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
	return c.File("views/list.html")
	// return c.HTML(http.StatusOK, fmt.Sprintf("%#v", sess))
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

func (this *WebController) List(c echo.Context) error {
	this.SetNoCache(c)
	sess, _ := session.Get("session", c)
	userID := sess.Values["oauthId"]
	if userID == "" {
		return c.Redirect(http.StatusTemporaryRedirect, "/")
	}
	todos, err := this.todoModel.List(userID.(string))
	if err != nil {
		return c.HTML(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, todos)
}

func (this *WebController) Pin(c echo.Context) error {
	this.SetNoCache(c)
	todo := new(Todo)
	if err := c.Bind(todo); err != nil {
		return c.HTML(http.StatusInternalServerError, err.Error())
	}
	if err := this.todoModel.Pin(*todo); err != nil {
		return c.HTML(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusOK)
}

func (this *WebController) Done(c echo.Context) error {
	this.SetNoCache(c)
	todo := new(Todo)
	if err := c.Bind(todo); err != nil {
		return c.HTML(http.StatusInternalServerError, err.Error())
	}
	log.Println(todo)
	if err := this.todoModel.Done(*todo); err != nil {
		return c.HTML(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusOK)
}

func (this *WebController) UserInfo(c echo.Context) error {
	this.SetNoCache(c)
	sess, _ := session.Get("session", c)
	data := map[string]string{
		"oauthName":    sess.Values["oauthName"].(string),
		"oauthPicture": sess.Values["oauthPicture"].(string),
	}

	return c.JSON(http.StatusOK, data)
}

func (this *WebController) Logout(c echo.Context) error {
	this.SetNoCache(c)
	sess, _ := session.Get("session", c)
	oauthToken := sess.Values["oauthToken"]
	if oauthToken != nil {
		err := this.oAuthService.Signout(oauthToken.(string))
		if err != nil {
			return c.HTML(http.StatusInternalServerError, err.Error())
		}
	}
	sess.Options.MaxAge = -1
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
		todoModel:    &todoModel,
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
	e.GET("/list", webController.List)
	e.POST("/pin", webController.Pin)
	e.POST("/done", webController.Done)
	e.GET("/user-info", webController.UserInfo)
	e.GET("/logout", webController.Logout)

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}
	fmt.Println("Runnign at :" + port)
	e.Logger.Fatal(e.Start(":" + port))
}

func (this *TodoBot) Remind(c echo.Context) error {
	userTodos, err := this.todoModel.Remind()
	log.Println(userTodos)
	if err != nil {
		return c.HTML(http.StatusInternalServerError, err.Error())
	}
	for userID, todos := range userTodos {
		message := "Hi there,\n"
		showDone := false
		remaining := 0
		for i, todo := range todos {
			if i == 0 && todo.Done {
				message += "Well done, you have no remaining tasks to be done :)\n"
			} else if i == 0 {
				message += "Tasks to be done:\n"
			}
			if !todo.Done {
				remaining++
			} else if todo.Done && !showDone {
				message += "Tasks completed:\n"
				showDone = true
			}
			if todo.Pin {
				message += "*** "
			} else {
				message += "    "
			}
			due := this.FormatDate(time.Now(), todo.Due)
			if !todo.Done && time.Now().After(todo.Due) {
				due += " (overdue)"
			}
			message += fmt.Sprintf("%v : %v\n", todo.Task, due)

		}
		if remaining != 0 {
			message += fmt.Sprintf("%d of %d remaining, just do it!", remaining, len(todos))
		}
		//Fork for massive API calls
		go this.PushMessage(userID, message)
	}
	return c.NoContent(http.StatusOK)
}

func (this *TodoBot) FormatDate(now time.Time, date time.Time) string {
	// Mon Jan 2 15:04:05 -0700 MST 2006
	dateText := date.Format("2006-01-02")
	timeText := date.Format("15:04")
	_, todayWeek := now.ISOWeek()
	_, dueWeek := date.ISOWeek()
	today := now.Format("2006-01-02")
	tomorrow := now.AddDate(0, 0, 1).Format("2006-01-02")
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")
	if dateText == today {
		// Today
		return "Today at " + timeText
	} else if dateText == tomorrow {
		// Tomorrow
		return "Tomorrow at " + timeText
	} else if dateText == yesterday {
		// Yesterday
		return "Yesterday at " + timeText
	} else if todayWeek == dueWeek && now.After(date) {
		// This week
		return "Last " + date.Format("Mon at 15:04")
	} else if todayWeek == dueWeek {
		// This week in the past
		return date.Format("Mon at 15:04")
	} else if dueWeek-todayWeek == 1 {
		// Next week
		return "Next " + date.Format("Mon at 15:04")
	}
	return date.Format("Mon 2 Jan 06 at 15:04")
}

func (this *TodoBot) PushMessage(userID string, message string) {
	if _, err := this.client.PushMessage(userID, linebot.NewTextMessage(message)).Do(); err != nil {
		log.Println(err)
	}
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
