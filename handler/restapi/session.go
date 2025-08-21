package restapi

import (
	"net/http"

	"github.com/dfcfw/goproxy/handler/session"
	"github.com/dfcfw/goproxy/handler/shipx"
	"github.com/xgfone/ship/v5"
)

func NewSession() *Session {
	return &Session{}
}

type Session struct{}

func (ses *Session) RegisterRoute(r *ship.RouteGroupBuilder) error {
	r.Route("/api/session/info").
		Data(shipx.NewRouteInfo("获取 session 信息").Logon().Map()).GET(ses.info)

	return nil
}

func (ses *Session) info(c *ship.Context) error {
	ret := session.FromMap(c.Data)

	return c.JSON(http.StatusOK, ret)
}
