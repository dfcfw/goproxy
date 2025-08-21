package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/dfcfw/goproxy/datalayer/model"
	"github.com/dfcfw/goproxy/datalayer/query"
)

func NewAdmin(qry *query.Query, log *slog.Logger) *Admin {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tbl := qry.Admin
	dao := tbl.WithContext(ctx)
	if cnt, _ := dao.Count(); cnt == 0 {
		_ = dao.Create(&model.Admin{JobNumber: "200858"})
	}

	return &Admin{
		qry: qry,
		log: log,
	}
}

type Admin struct {
	qry *query.Query
	log *slog.Logger
}

func (adm *Admin) List(ctx context.Context) ([]*model.Admin, error) {
	tbl := adm.qry.Admin
	dao := tbl.WithContext(ctx)

	return dao.Find()
}

func (adm *Admin) Create(ctx context.Context, jobNumber string) error {
	tbl := adm.qry.Admin
	dao := tbl.WithContext(ctx)
	dat := &model.Admin{JobNumber: jobNumber}

	return dao.Create(dat)
}

func (adm *Admin) Delete(ctx context.Context, jobNumber string) error {
	tbl := adm.qry.Admin
	dao := tbl.WithContext(ctx)
	_, err := dao.Where(tbl.JobNumber.Eq(jobNumber)).Delete()

	return err
}
