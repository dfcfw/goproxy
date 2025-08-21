package restapi

import (
	"github.com/dfcfw/goproxy/handler/shipx"
	"github.com/xgfone/ship/v5"
)

func NewProxy(dir string) *Proxy {
	return &Proxy{
		dir: dir,
	}
}

type Proxy struct {
	dir string
}

func (prx *Proxy) RegisterRoute(r *ship.RouteGroupBuilder) error {
	r.Route("/private").
		Data(shipx.NewRouteInfo("模块代理").UsePAT().Map()).Static(prx.dir)

	return nil
}
