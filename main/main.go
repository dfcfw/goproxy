package main

import (
	"log/slog"
	"net/http"

	"github.com/dfcfw/goproxy/business/service"
	"github.com/dfcfw/goproxy/handler/restapi"
	"github.com/xgfone/ship/v5"
)

func main() {
	moddir := "/Users/wang/Documents/gocode/pkg/mod/cache/download/"
	log := slog.Default()
	sh := ship.Default()

	gomodSvc := service.NewGomod(moddir, log)
	gomodAPI := restapi.NewGomod(gomodSvc)
	apiRGB := sh.Group("/api")
	apiRGB.Route("/gomod/browse").GET(gomodAPI.Browse)

	http.ListenAndServe(":65432", sh)
}
