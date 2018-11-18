package service

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
)

type SessionService interface {
	Get(c echo.Context, name string) interface{}
	Set(c echo.Context, name string, value interface{})
	Destroy(c echo.Context)
}

type CookieSessionService struct {
}

func (this *CookieSessionService) Get(c echo.Context, name string) interface{} {
	sess, _ := session.Get("session", c)
	return sess.Values[name]
}

func (this *CookieSessionService) Destroy(c echo.Context) {
	sess, _ := session.Get("session", c)
	sess.Options.MaxAge = -1
	sess.Save(c.Request(), c.Response())
}

func (this *CookieSessionService) Set(c echo.Context, name string, value interface{}) {
	sess, _ := session.Get("session", c)
	sess.Values[name] = value
	sess.Save(c.Request(), c.Response())
}
