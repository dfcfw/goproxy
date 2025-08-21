package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"log/slog"
	"time"

	"github.com/dfcfw/goproxy/contract/errcode"
	"github.com/dfcfw/goproxy/contract/request"
	"github.com/dfcfw/goproxy/datalayer/model"
	"github.com/dfcfw/goproxy/datalayer/query"
)

func NewAccessToken(qry *query.Query, log *slog.Logger) *AccessToken {
	return &AccessToken{
		qry:   qry,
		log:   log,
		limit: 10,
	}
}

type AccessToken struct {
	qry   *query.Query
	log   *slog.Logger
	limit int64
}

func (pat *AccessToken) List(ctx context.Context, jobNumber string) ([]*model.AccessToken, error) {
	tbl := pat.qry.AccessToken
	dao := tbl.WithContext(ctx)

	return dao.Omit(tbl.Token).Where(tbl.JobNumber.Eq(jobNumber)).Find()
}

func (pat *AccessToken) Create(ctx context.Context, jobNumber string, req *request.AccessTokenCreate) (*model.AccessToken, error) {
	now := time.Now()
	buf := make([]byte, 28)
	binary.LittleEndian.PutUint64(buf, uint64(now.UnixNano()))
	_, _ = rand.Read(buf[8:])
	token := "pat_" + base64.RawURLEncoding.EncodeToString(buf)

	dat := &model.AccessToken{
		Name:      req.Name,
		JobNumber: jobNumber,
		Token:     token,
		ExpiredAt: req.ExpiredAt,
	}

	err := pat.qry.Transaction(func(tx *query.Query) error {
		tbl := tx.AccessToken
		dao := tbl.WithContext(ctx)
		cnt, _ := dao.Where(tbl.JobNumber.Eq(jobNumber)).Count()
		if cnt >= pat.limit {
			return errcode.FmtPATLimited.Fmt(pat.limit)
		}

		return dao.Create(dat)
	})
	if err != nil {
		return nil, err
	}

	return dat, nil
}

func (pat *AccessToken) Delete(ctx context.Context, jobNumber, name string) error {
	tbl := pat.qry.AccessToken
	dao := tbl.WithContext(ctx)
	ret, err := dao.Where(tbl.JobNumber.Eq(jobNumber), tbl.Name.Eq(name)).Delete()
	if err != nil {
		return err
	}
	if ret.RowsAffected == 0 {
		return errcode.ErrDataNotExists
	}

	return nil
}

func (pat *AccessToken) Exists(ctx context.Context, jobNumber, name string) bool {
	tbl := pat.qry.AccessToken
	dao := tbl.WithContext(ctx)
	cnt, _ := dao.Where(tbl.JobNumber.Eq(jobNumber), tbl.Name.Eq(name)).Count()

	return cnt != 0
}
