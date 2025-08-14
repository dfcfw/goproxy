package restapi

import (
	"archive/zip"
	"bytes"
	"net/http"

	"github.com/dfcfw/goproxy/business/service"
	"github.com/dfcfw/goproxy/contract/request"
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

func (gmd *Gomod) Walk(c *ship.Context) error {
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

func (gmd *Gomod) Stat(c *ship.Context) error {
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

func (gmd *Gomod) Sniff(c *ship.Context) error {
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

func (gmd *Gomod) Upload(c *ship.Context) error {
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

func (gmd *Gomod) Format(c *ship.Context) error {
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
