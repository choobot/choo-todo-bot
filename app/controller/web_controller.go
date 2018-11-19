package controller

import (
	"go/build"
	"log"
	"net/http"

	"github.com/choobot/choo-todo-bot/app/model"
	"github.com/choobot/choo-todo-bot/app/service"
	"github.com/labstack/echo"
	"golang.org/x/oauth2"
)

type WebController struct {
	OAuthService   service.OAuthService
	JwtService     service.JwtService
	TodoModel      model.TodoModel
	SessionService service.SessionService
}

func (this *WebController) SetNoCache(c echo.Context) {
	c.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Response().Header().Set("Pragma", "no-cache")
	c.Response().Header().Set("Expires", "0")
}

func (this *WebController) Index(c echo.Context) error {
	this.SetNoCache(c)
	oauthToken := this.SessionService.Get(c, "oauthToken")
	if oauthToken == nil {
		return c.File(build.Default.GOPATH + "/src/github.com/choobot/choo-todo-bot/app/views/login.html")
	}
	return c.File(build.Default.GOPATH + "/src/github.com/choobot/choo-todo-bot/app/views/list.html")
}

func (this *WebController) Login(c echo.Context) error {
	this.SetNoCache(c)
	oauthState := this.OAuthService.GenerateOAuthState()
	this.SessionService.Set(c, "oauthState", oauthState)
	url := this.OAuthService.OAuthConfig().AuthCodeURL(oauthState, oauth2.AccessTypeOnline)
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

func (this *WebController) Auth(c echo.Context) error {
	this.SetNoCache(c)
	oauthState := this.SessionService.Get(c, "oauthState")

	state := c.QueryParam("state")
	if oauthState != "" && state != oauthState {
		log.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthState, state)
		return c.Redirect(http.StatusTemporaryRedirect, "/")
	}
	code := c.QueryParam("code")
	oauthToken, err := this.OAuthService.OAuthConfig().Exchange(oauth2.NoContext, code)
	if err != nil {
		return c.HTML(http.StatusInternalServerError, err.Error())
	}

	idToken, err := this.JwtService.ExtractIdToken(oauthToken.Extra("id_token").(string))
	if err != nil {
		return c.HTML(http.StatusInternalServerError, err.Error())
	}
	this.SessionService.Set(c, "oauthToken", oauthToken.AccessToken)
	this.SessionService.Set(c, "oauthId", idToken.JWT.Subject)
	this.SessionService.Set(c, "oauthName", idToken.Name)
	this.SessionService.Set(c, "oauthPicture", idToken.Picture)
	return c.Redirect(http.StatusTemporaryRedirect, "/")
}

func (this *WebController) List(c echo.Context) error {
	this.SetNoCache(c)
	userID := this.SessionService.Get(c, "oauthId")
	if userID == nil {
		return c.HTML(http.StatusInternalServerError, "user not found")
	}
	todos, err := this.TodoModel.List(userID.(string))
	if err != nil {
		return c.HTML(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, todos)
}

func (this *WebController) Pin(c echo.Context) error {
	this.SetNoCache(c)
	todo := new(model.Todo)
	if err := c.Bind(todo); err != nil {
		return c.HTML(http.StatusInternalServerError, err.Error())
	}
	if err := this.TodoModel.Pin(*todo); err != nil {
		return c.HTML(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusOK)
}

func (this *WebController) Done(c echo.Context) error {
	this.SetNoCache(c)
	todo := new(model.Todo)
	if err := c.Bind(todo); err != nil {
		return c.HTML(http.StatusInternalServerError, err.Error())
	}
	if err := this.TodoModel.Done(*todo); err != nil {
		return c.HTML(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusOK)
}

func (this *WebController) UserInfo(c echo.Context) error {
	this.SetNoCache(c)
	oauthName := this.SessionService.Get(c, "oauthName")
	oauthPicture := this.SessionService.Get(c, "oauthPicture")
	if oauthName == nil || oauthPicture == nil {
		return c.HTML(http.StatusInternalServerError, "user not found")
	}
	data := map[string]string{
		"oauthName":    oauthName.(string),
		"oauthPicture": oauthPicture.(string),
	}

	return c.JSON(http.StatusOK, data)
}

func (this *WebController) Logout(c echo.Context) error {
	this.SetNoCache(c)
	oauthToken := this.SessionService.Get(c, "oauthToken")
	if oauthToken != nil {
		err := this.OAuthService.Signout(oauthToken.(string))
		if err != nil {
			return c.HTML(http.StatusInternalServerError, err.Error())
		}
	}
	this.SessionService.Destroy(c)
	return c.Redirect(http.StatusTemporaryRedirect, "/")
}

func (this *WebController) Edit(c echo.Context) error {
	this.SetNoCache(c)
	todo := new(model.Todo)
	if err := c.Bind(todo); err != nil {
		return c.HTML(http.StatusInternalServerError, err.Error())
	}
	if err := this.TodoModel.Edit(*todo); err != nil {
		return c.HTML(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusOK)
}

func (this *WebController) Delete(c echo.Context) error {
	this.SetNoCache(c)
	todo := new(model.Todo)
	if err := c.Bind(todo); err != nil {
		return c.HTML(http.StatusInternalServerError, err.Error())
	}
	if err := this.TodoModel.Delete(*todo); err != nil {
		return c.HTML(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusOK)
}
