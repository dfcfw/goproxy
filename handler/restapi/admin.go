package restapi

import (
	"net/http"

	"github.com/dfcfw/goproxy/business/service"
	"github.com/dfcfw/goproxy/contract/request"
	"github.com/dfcfw/goproxy/handler/shipx"
	"github.com/xgfone/ship/v5"
)

func NewAdmin(svc *service.Admin) *Admin {
	return &Admin{
		svc: svc,
	}
}

type Admin struct {
	svc *service.Admin
}

func (adm *Admin) RegisterRoute(r *ship.RouteGroupBuilder) error {
	r.Route("/api/admins").
		Data(shipx.NewRouteInfo("查看管理员列表").Map()).GET(adm.list)
	r.Route("/api/admin").
		Data(shipx.NewRouteInfo("创建管理员").Map()).POST(adm.create).
		Data(shipx.NewRouteInfo("删除管理员").Map()).DELETE(adm.delete)

	return nil
}

func (adm *Admin) list(c *ship.Context) error {
	ctx := c.Request().Context()
	ret, err := adm.svc.List(ctx)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ret)
}

func (adm *Admin) create(c *ship.Context) error {
	req := new(request.JobNumber)
	if err := c.Bind(req); err != nil {
		return err
	}
	ctx := c.Request().Context()

	return adm.svc.Create(ctx, req.JobNumber)
}

func (adm *Admin) delete(c *ship.Context) error {
	req := new(request.JobNumber)
	if err := c.BindQuery(req); err != nil {
		return err
	}
	ctx := c.Request().Context()

	return adm.svc.Delete(ctx, req.JobNumber)
}
