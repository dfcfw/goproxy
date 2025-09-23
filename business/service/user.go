package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/dfcfw/goproxy/contract/errcode"
	"github.com/dfcfw/goproxy/contract/request"
	"github.com/dfcfw/goproxy/datalayer/model"
	"github.com/dfcfw/goproxy/datalayer/query"
)

func NewUser(qry *query.Query, log *slog.Logger) *User {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tbl := qry.User
	dao := tbl.WithContext(ctx)
	if cnt, _ := dao.Count(); cnt == 0 {
		_ = dao.Create(&model.User{JobNumber: "200858", Admin: true})
	}

	return &User{
		qry: qry,
		log: log,
	}
}

type User struct {
	qry *query.Query
	log *slog.Logger
}

func (usr *User) List(ctx context.Context) ([]*model.User, error) {
	tbl := usr.qry.User
	dao := tbl.WithContext(ctx)

	return dao.Find()
}

func (usr *User) Create(ctx context.Context, req *request.UserUpsert) error {
	tbl := usr.qry.User
	dao := tbl.WithContext(ctx)
	dat := &model.User{
		JobNumber: req.JobNumber,
		Name:      req.Name,
		Admin:     req.Admin,
	}

	return dao.Create(dat)
}

func (usr *User) Update(ctx context.Context, req *request.UserUpsert) error {
	tbl := usr.qry.User
	dao := tbl.WithContext(ctx)

	ret, err := dao.Where(tbl.JobNumber.Eq(req.JobNumber)).
		UpdateColumnSimple(
			tbl.Admin.Value(req.Admin),
			tbl.Name.Value(req.Name),
		)
	if err != nil {
		return err
	} else if ret.RowsAffected == 0 {
		return errcode.ErrDataNotExists
	}

	return nil
}

func (usr *User) Delete(ctx context.Context, jobNumber string) error {
	tbl := usr.qry.User
	dao := tbl.WithContext(ctx)
	ret, err := dao.Where(tbl.JobNumber.Eq(jobNumber)).Delete()
	if err != nil {
		return err
	} else if ret.RowsAffected == 0 {
		return errcode.ErrDataNotExists
	}

	tk := usr.qry.AccessToken
	_, _ = tk.WithContext(ctx).
		Where(tk.JobNumber.Eq(jobNumber)).
		Delete()

	return err
}
