package session

import (
	"context"
	"log/slog"
	"time"

	"github.com/dfcfw/goproxy/business/jwtoken"
	"github.com/dfcfw/goproxy/datalayer/model"
	"github.com/dfcfw/goproxy/datalayer/query"
	"github.com/dfcfw/goproxy/integration/casauth"
)

type Validator interface {
	ValidPAT(ctx context.Context, token string) (*Userinfo, error)
	ValidCAS(ctx context.Context, name, passwd string) (*Userinfo, error)
	ValidJWT(ctx context.Context, token string) (*Userinfo, error)
	SignJWT(jobNumber string, period time.Duration) (string, error)
}

func NewValid(qry *query.Query, cas casauth.Client, tok *jwtoken.Issue, log *slog.Logger) Validator {
	return &identValid{
		qry: qry,
		cas: cas,
		tok: tok,
		log: log,
	}
}

type identValid struct {
	qry *query.Query
	cas casauth.Client
	tok *jwtoken.Issue
	log *slog.Logger
}

func (idt *identValid) ValidPAT(ctx context.Context, token string) (*Userinfo, error) {
	if token == "" {
		return nil, nil
	}

	now := time.Now()
	tbl := idt.qry.AccessToken
	dao := tbl.WithContext(ctx)

	dat, err := dao.Where(tbl.Token.Eq(token)).First()
	if err != nil {
		return nil, err
	}

	// exp.IsZero() 表示永不过期
	exp := dat.ExpiredAt
	if !exp.IsZero() && exp.Before(now) {
		return nil, nil
	}

	jobNumber := dat.JobNumber
	user, err := idt.valid(ctx, jobNumber)
	if err != nil {
		return nil, err
	}
	info := &Userinfo{JobNumber: jobNumber, Admin: user.Admin}

	return info, nil
}

func (idt *identValid) ValidCAS(ctx context.Context, name, passwd string) (*Userinfo, error) {
	user, err := idt.valid(ctx, name)
	if err != nil {
		return nil, err
	}

	if err = idt.cas.Auth(ctx, name, passwd); err != nil {
		return nil, err
	}

	info := &Userinfo{JobNumber: name, Admin: user.Admin}

	return info, nil
}

func (idt *identValid) ValidJWT(ctx context.Context, token string) (*Userinfo, error) {
	claim, err := idt.tok.Valid(token)
	if err != nil {
		return nil, err
	}

	jobNumber := claim.JobNumber
	user, err := idt.valid(ctx, jobNumber)
	if err != nil {
		return nil, err
	}
	info := &Userinfo{JobNumber: jobNumber, Admin: user.Admin}

	return info, nil
}

func (idt *identValid) SignJWT(jobNumber string, period time.Duration) (string, error) {
	return idt.tok.Sign(jobNumber, period)
}

func (idt *identValid) valid(ctx context.Context, jobNumber string) (*model.User, error) {
	tbl := idt.qry.User
	dao := tbl.WithContext(ctx)

	return dao.Where(tbl.JobNumber.Eq(jobNumber)).First()
}
