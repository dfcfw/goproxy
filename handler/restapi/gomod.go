package restapi

import (
	"archive/zip"
	"bytes"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/dfcfw/goproxy/business/service"
	"github.com/dfcfw/goproxy/contract/request"
	"github.com/dfcfw/goproxy/handler/shipx"
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

func (gmd *Gomod) RegisterRoute(r *ship.RouteGroupBuilder) error {
	r.Route("/api/gomod/walk").
		Data(shipx.NewRouteInfo("查看目录").Logon().Map()).GET(gmd.walk)
	r.Route("/api/gomod/stat").
		Data(shipx.NewRouteInfo("获取特定版本文件列表").Logon().Map()).GET(gmd.stat)
	r.Route("/api/gomod/file").
		Data(shipx.NewRouteInfo("下载文件").Logon().Map()).GET(gmd.file)
	r.Route("/api/gomod/sniff").
		Data(shipx.NewRouteInfo("探测模块版本信息").Map()).PUT(gmd.sniff)
	r.Route("/api/gomod/upload").
		Data(shipx.NewRouteInfo("上传模块文件").Map()).PUT(gmd.upload)
	r.Route("/api/gomod/format").
		Data(shipx.NewRouteInfo("格式转换").Map()).PUT(gmd.format)

	return nil
}

func (gmd *Gomod) walk(c *ship.Context) error {
	req := new(request.GomodWalk)
	if err := c.BindQuery(req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	ret, err := gmd.svc.Walk(ctx, req.Path)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ret)
}

func (gmd *Gomod) stat(c *ship.Context) error {
	req := new(request.GomodStat)
	if err := c.BindQuery(req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	ret, err := gmd.svc.Stat(ctx, req.Path, req.Version)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ret)
}

func (gmd *Gomod) sniff(c *ship.Context) error {
	req := new(request.GomodSniff)
	if err := c.Bind(req); err != nil {
		return err
	}

	size := req.File.Size
	file, err := req.File.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	ret, _ := gmd.svc.Sniff(file, size)

	return c.JSON(http.StatusOK, ret)
}

func (gmd *Gomod) upload(c *ship.Context) error {
	req := new(request.GomodUpload)
	if err := c.Bind(req); err != nil {
		return err
	}

	file, err := req.File.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	return gmd.svc.Upload(file, req.Path, req.Version)
}

func (gmd *Gomod) format(c *ship.Context) error {
	req := new(request.GomodUpload)
	if err := c.Bind(req); err != nil {
		return err
	}

	size := req.File.Size
	file, err := req.File.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	rd, err := zip.NewReader(file, size)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	err = gmd.svc.Format(buf, rd, req.Path, req.Version)
	if err != nil {
		return err
	}

	return c.Stream(http.StatusOK, "application/zip", buf)
}

func (gmd *Gomod) file(c *ship.Context) error {
	req := new(request.GomodFile)
	if err := c.BindQuery(req); err != nil {
		return err
	}

	modpath, name := req.Path, req.Name
	ext := filepath.Ext(name)
	ct := mime.TypeByExtension(ext)
	if ct == "" {
		ct = ship.MIMEOctetStream
	}

	file, err := gmd.svc.Open(modpath, name)
	if err != nil {
		return err
	}
	defer file.Close()

	disposition := mime.FormatMediaType("attachment", map[string]string{"filename": name})
	c.SetRespHeader(ship.HeaderContentDisposition, disposition)
	if inf, _ := file.Stat(); inf != nil {
		c.SetRespHeader(ship.HeaderContentLength, strconv.FormatInt(inf.Size(), 10))
	}

	return c.Stream(http.StatusOK, ct, file)
}
