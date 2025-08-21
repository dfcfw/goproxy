package shipx

import "github.com/xgfone/ship/v5"

type RouteRegister interface {
	RegisterRoute(*ship.RouteGroupBuilder) error
}

func RegisterRoutes(rgb *ship.RouteGroupBuilder, rrs []RouteRegister) error {
	for _, rr := range rrs {
		if err := rr.RegisterRoute(rgb); err != nil {
			return err
		}
	}

	return nil
}
