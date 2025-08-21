package errcode

import "github.com/xgfone/ship/v5"

var (
	ErrDataNotExists = ship.ErrBadRequest.Newf("数据不存在")
	ErrNotFound      = ship.ErrNotFound.Newf("资源不存在")
)

var FmtPATLimited = stringError("token 不得超过 %d 个")

type Formatter interface {
	Fmt(v ...any) error
}

type stringError string

func (s stringError) Fmt(v ...any) error {
	return ship.ErrBadRequest.Newf(string(s), v...)
}
