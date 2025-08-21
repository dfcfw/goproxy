package errcode

import "github.com/xgfone/ship/v5"

var (
	ErrNotFound       = ship.ErrNotFound.Newf("资源不存在")
	ErrUnauthorized   = ship.ErrUnauthorized.Newf("请先登录")
	ErrInternalServer = ship.ErrInternalServerError.Newf("内部错误")
	ErrDataNotExists  = ship.ErrBadRequest.Newf("数据不存在")
	ErrInvalidToken   = ship.ErrBadRequest.Newf("无效的 Token")
)

var FmtPATLimited = stringError("token 不得超过 %d 个")

type Formatter interface {
	Fmt(v ...any) error
}

type stringError string

func (s stringError) Fmt(v ...any) error {
	return ship.ErrBadRequest.Newf(string(s), v...)
}
