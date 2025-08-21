package restapi

import (
	"net/http"

	"github.com/dfcfw/goproxy/business/service"
	"github.com/dfcfw/goproxy/contract/request"
	"github.com/dfcfw/goproxy/handler/session"
	"github.com/dfcfw/goproxy/handler/shipx"
	"github.com/xgfone/ship/v5"
)

func NewAccessToken(svc *service.AccessToken) *AccessToken {
	return &AccessToken{svc: svc}
}

type AccessToken struct {
	svc *service.AccessToken
}

func (pat *AccessToken) RegisterRoute(r *ship.RouteGroupBuilder) error {
	r.Route("/api/access-tokens").
		Data(shipx.NewRouteInfo("查看 PAT 列表").Logon().Map()).GET(pat.list)
	r.Route("/api/access-token").
		Data(shipx.NewRouteInfo("创建 PAT").Logon().Map()).POST(pat.create).
		Data(shipx.NewRouteInfo("删除 PAT").Logon().Map()).DELETE(pat.delete)
	r.Route("/api/access-token/valid").
		Data(shipx.NewRouteInfo("检查 PAT 名字是否可用").Logon().Map()).GET(pat.valid)

	return nil
}

func (pat *AccessToken) list(c *ship.Context) error {
	ctx := c.Request().Context()
	sess := session.FromMap(c.Data)

	ret, err := pat.svc.List(ctx, sess.ID())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ret)
}

func (pat *AccessToken) create(c *ship.Context) error {
	req := new(request.AccessTokenCreate)
	if err := c.Bind(req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	sess := session.FromMap(c.Data)

	ret, err := pat.svc.Create(ctx, sess.ID(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ret)
}

func (pat *AccessToken) delete(c *ship.Context) error {
	req := new(request.Named)
	if err := c.BindQuery(req); err != nil {
		return err
	}
	ctx := c.Request().Context()
	sess := session.FromMap(c.Data)

	return pat.svc.Delete(ctx, sess.ID(), req.Name)
}

func (pat *AccessToken) valid(c *ship.Context) error {
	req := new(request.Named)
	if err := c.BindQuery(req); err != nil {
		return err
	}
	ctx := c.Request().Context()
	sess := session.FromMap(c.Data)
	exists := pat.svc.Exists(ctx, sess.ID(), req.Name)

	return c.JSON(http.StatusOK, map[string]bool{"succeed": !exists})
}
