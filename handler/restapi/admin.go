package restapi

import (
	"net/http"

	"github.com/dfcfw/goproxy/business/service"
	"github.com/dfcfw/goproxy/contract/request"
	"github.com/dfcfw/goproxy/handler/shipx"
	"github.com/xgfone/ship/v5"
)

func NewUser(svc *service.User) *User {
	return &User{
		svc: svc,
	}
}

type User struct {
	svc *service.User
}

func (usr *User) RegisterRoute(r *ship.RouteGroupBuilder) error {
	r.Route("/api/users").
		Data(shipx.NewRouteInfo("查看用户列表").Map()).GET(usr.list)
	r.Route("/api/user").
		Data(shipx.NewRouteInfo("创建用户").Map()).POST(usr.create).
		Data(shipx.NewRouteInfo("修改用户").Map()).PUT(usr.update).
		Data(shipx.NewRouteInfo("删除用户").Map()).DELETE(usr.delete)

	return nil
}

func (usr *User) list(c *ship.Context) error {
	ctx := c.Request().Context()
	ret, err := usr.svc.List(ctx)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ret)
}

func (usr *User) create(c *ship.Context) error {
	req := new(request.UserUpsert)
	if err := c.Bind(req); err != nil {
		return err
	}
	ctx := c.Request().Context()

	return usr.svc.Create(ctx, req)
}

func (usr *User) update(c *ship.Context) error {
	req := new(request.UserUpsert)
	if err := c.Bind(req); err != nil {
		return err
	}
	ctx := c.Request().Context()

	return usr.svc.Update(ctx, req)
}

func (usr *User) delete(c *ship.Context) error {
	req := new(request.JobNumber)
	if err := c.BindQuery(req); err != nil {
		return err
	}
	ctx := c.Request().Context()

	return usr.svc.Delete(ctx, req.JobNumber)
}
