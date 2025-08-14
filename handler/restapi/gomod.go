package restapi

import (
	"github.com/dfcfw/goproxy/business/service"
	"github.com/xgfone/ship/v5"
)

func NewGomod(svc *service.Gomod) *Gomod {
	return &Gomod{
		svc: svc,
	}
}

type Gomod struct {
	svc *service.Gomod
}

func (gmd *Gomod) Browse(c *ship.Context) error {
	node := c.Query("node")
	ctx := c.Request().Context()
	_ = gmd.svc.Browse(ctx, node)

	return nil
}
