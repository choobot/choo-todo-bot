package main

import (
	"go/build"
	"log"
	"net/http"
	"os"

	"github.com/choobot/choo-todo-bot/app/bot"
	"github.com/choobot/choo-todo-bot/app/controller"
	"github.com/choobot/choo-todo-bot/app/model"
	"github.com/choobot/choo-todo-bot/app/service"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/middleware"
	"github.com/line/line-bot-sdk-go/linebot"
)

func main() {
	client, err := linebot.New(os.Getenv("LINE_BOT_SECRET"), os.Getenv("LINE_BOT_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	todoModel := model.NewTodoMySqlModel()

	bot := &bot.TodoBot{
		TodoModel: &todoModel,
		Client:    client,
	}
	oAuthSerivce := service.NewLineOAuthService()
	jwtService := service.NewLineJwtService()
	webController := controller.WebController{
		OAuthService:   &oAuthSerivce,
		JwtService:     &jwtService,
		TodoModel:      &todoModel,
		SessionService: &service.CookieSessionService{},
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("choo-todo-bot"))))

	// Routes
	e.POST("/callback", func(c echo.Context) error {
		events, err := client.ParseRequest(c.Request())
		if err != nil {
			if err == linebot.ErrInvalidSignature {
				return c.NoContent(http.StatusBadRequest)
			} else {
				return c.HTML(http.StatusInternalServerError, err.Error())
			}
		}
		if err := bot.Response(events); err != nil {
			log.Println(err)
			return c.HTML(http.StatusInternalServerError, err.Error())
		}
		return c.NoContent(http.StatusOK)
	})
	e.GET("/remind", func(c echo.Context) error {
		err := bot.Remind()
		if err != nil {
			return c.HTML(http.StatusInternalServerError, err.Error())
		}
		return c.NoContent(http.StatusOK)
	})
	e.Static("/", build.Default.GOPATH+"/src/github.com/choobot/choo-todo-bot/app/assets")
	e.GET("/", webController.Index)
	e.GET("/login", webController.Login)
	e.GET("/auth", webController.Auth)
	e.GET("/list", webController.List)
	e.POST("/pin", webController.Pin)
	e.POST("/done", webController.Done)
	e.GET("/user-info", webController.UserInfo)
	e.GET("/logout", webController.Logout)
	e.POST("/edit", webController.Edit)
	e.POST("/delete", webController.Delete)

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
