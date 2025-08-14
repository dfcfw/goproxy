package main

import (
	"log/slog"
	"net/http"

	"github.com/dfcfw/goproxy/business/service"
	"github.com/dfcfw/goproxy/handler/restapi"
	"github.com/xgfone/ship/v5"
)

func main() {
	// moddir := "/Users/wang/Documents/gocode/pkg/mod/cache/download/"
	moddir := "resources/mod/"
	log := slog.Default()
	sh := ship.Default()
	sh.HandleError = func(c *ship.Context, err error) {
		c.JSON(http.StatusBadRequest, map[string]string{"detail": err.Error()})
	}

	rootRGB := sh.Group("/")
	rootRGB.Route("/").Static("resources/static/root/")
	rootRGB.Route("/oas3").Static("resources/static/oas3/")
	rootRGB.Route("/private").Static(moddir)

	gomodSvc := service.NewGomod(moddir, log)
	gomodAPI := restapi.NewGomod(gomodSvc)
	apiRGB := rootRGB.Group("/api")
	apiRGB.Route("/gomod/walk").GET(gomodAPI.Walk)
	apiRGB.Route("/gomod/stat").GET(gomodAPI.Stat)
	apiRGB.Route("/gomod/sniff").PUT(gomodAPI.Sniff)
	apiRGB.Route("/gomod/upload").PUT(gomodAPI.Upload)
	apiRGB.Route("/gomod/format").PUT(gomodAPI.Format)

	_ = http.ListenAndServe(":65432", sh)
}
