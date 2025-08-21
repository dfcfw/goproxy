package launch

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/dfcfw/goproxy/business/jwtoken"
	"github.com/dfcfw/goproxy/business/service"
	"github.com/dfcfw/goproxy/config"
	"github.com/dfcfw/goproxy/datalayer/model"
	"github.com/dfcfw/goproxy/datalayer/query"
	"github.com/dfcfw/goproxy/handler/middle"
	"github.com/dfcfw/goproxy/handler/restapi"
	"github.com/dfcfw/goproxy/handler/session"
	"github.com/dfcfw/goproxy/handler/shipx"
	"github.com/dfcfw/goproxy/integration/casauth"
	"github.com/dfcfw/goproxy/library/httpx"
	"github.com/dfcfw/goproxy/library/jsonc"
	"github.com/glebarez/sqlite"
	"github.com/xgfone/ship/v5"
	"gorm.io/gorm"
)

func Run(ctx context.Context, cfgFile string) error {
	const safeSize = 1 << 20
	cfg := new(config.Config)
	if err := jsonc.ReadFile(cfgFile, cfg, safeSize); err != nil { // 读取主配置文件
		return err
	}

	return Exec(ctx, cfg)
}

//goland:noinspection GoUnhandledErrorResult
func Exec(ctx context.Context, cfg *config.Config) error {
	log := slog.Default()
	srvCfg, dbCfg := cfg.Server, cfg.Database
	db, err := gorm.Open(sqlite.Open(dbCfg.DSN))
	if err != nil {
		return err
	}
	if err = db.AutoMigrate(model.All()...); err != nil {
		return err
	}
	qry := query.Use(db)

	httpClient := httpx.NewClient(http.DefaultClient)
	casCfg := casauth.StringURL(srvCfg.CAS)
	casClient := casauth.NewClient(casCfg, httpClient, log)

	const moddir = "resources/mod/"
	userSvc := service.NewUser(qry, log)
	accessTokenSvc := service.NewAccessToken(qry, log)
	gomodSvc := service.NewGomod(moddir, log)

	jwtIssue := jwtoken.NewIssue(nil, log)
	sessValid := session.NewValid(qry, casClient, jwtIssue, log)
	authMiddle := middle.NewAuth(sessValid)

	restAPIs := []shipx.RouteRegister{
		restapi.NewAccessToken(accessTokenSvc),
		restapi.NewGomod(gomodSvc),
		restapi.NewSession(),
		restapi.NewUser(userSvc),
		restapi.NewProxy(moddir),
	}

	shipHTTP := ship.Default()
	shipHTTP.NotFound = shipx.NotFound
	shipHTTP.HandleError = shipx.HandleError
	rootRGB := shipHTTP.Group("/")
	for k, v := range srvCfg.Static {
		if k != "" && v != "" {
			rootRGB.Route(k).Static(v)
		}
	}

	restRBG := rootRGB.Use(authMiddle)
	if err = shipx.RegisterRoutes(restRBG, restAPIs); err != nil {
		return err
	}

	errs := make(chan error, 1)
	srv := &http.Server{
		Handler: shipHTTP,
		Addr:    srvCfg.Addr,
	}
	go listenAndServe(errs, srv)
	select {
	case err = <-errs:
	case <-ctx.Done():
	}
	_ = srv.Close()

	return err
}

func listenAndServe(errs chan error, srv *http.Server) {
	errs <- srv.ListenAndServe()
}
