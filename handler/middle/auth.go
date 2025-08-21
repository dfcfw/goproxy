package middle

import (
	"net/http"
	"net/netip"
	"strings"
	"time"

	"github.com/dfcfw/goproxy/handler/session"
	"github.com/dfcfw/goproxy/handler/shipx"
	"github.com/xgfone/ship/v5"
	"golang.org/x/net/publicsuffix"
)

func NewAuth(valid session.Validator) ship.Middleware {
	atm := &authMiddle{
		valid:      valid,
		period:     time.Hour,
		cookieName: "goproxy-bearer",
	}

	return atm.call
}

type authMiddle struct {
	valid      session.Validator
	period     time.Duration
	cookieName string // JWT 存放的 Cookie Name
}

func (atm *authMiddle) call(h ship.Handler) ship.Handler {
	return func(c *ship.Context) error {
		// 获取路由信息，如果该路由没有配置相关信息，以最小化权限原则：
		// 没有配置路由信息的接口，只能管理员访问。
		sessKey := session.Key.String()
		info := shipx.DetectRouteInfo(c.Route.Data)
		if info == nil {
			return ship.ErrInternalServerError
		}

		r := c.Request()
		ctx := r.Context()

		perm := info.Perm()
		if perm.Anonymous { // 允许匿名访问，则无需做任何校验
			return h(c)
		}
		if perm.UsePAT { // 如果使用 PAT 认证
			name, _, _ := r.BasicAuth()
			if sess, _ := atm.valid.ValidPAT(ctx, name); sess != nil {
				c.Data[sessKey] = sess
				return h(c)
			}

			return atm.needAuth(c)
		}

		sess := atm.parseUser(c)
		if sess == nil {
			return atm.needAuth(c)
		}
		if !perm.Logon && !sess.Admin {
			return ship.ErrForbidden
		}

		c.Data[sessKey] = sess

		return h(c)
	}
}

func (atm *authMiddle) parseUser(c *ship.Context) *session.Userinfo {
	// 先从 cookie 中的解析 jwt。
	r := c.Request()
	ctx := r.Context()
	if cookie, _ := r.Cookie(atm.cookieName); cookie != nil {
		info, err := atm.valid.ValidJWT(ctx, cookie.Value)
		if err == nil {
			return info
		}
	}

	jobNumber, passwd, ok := r.BasicAuth()
	if !ok || jobNumber == "" || passwd == "" {
		return nil
	}

	info, err := atm.valid.ValidCAS(ctx, jobNumber, passwd)
	if err != nil {
		return nil
	}
	bearer, err := atm.valid.SignJWT(jobNumber, atm.period)
	if err != nil {
		return nil
	}

	expiredAt := time.Now().Add(atm.period)
	domain := atm.cookieDomain(c.Host())
	cookie := &http.Cookie{
		Name:     atm.cookieName,
		Value:    bearer,
		Expires:  expiredAt,
		Domain:   domain,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
	}
	c.SetCookie(cookie)

	return info
}

func (atm *authMiddle) cookieDomain(host string) string {
	if addr, err := netip.ParseAddr(host); err == nil {
		return addr.String()
	}

	suffix, _ := publicsuffix.PublicSuffix(host)
	if suffix == "" {
		return ""
	}

	before, _ := strings.CutSuffix(host, suffix)
	splits := strings.Split(strings.Trim(before, "."), ".")
	if size := len(splits); size != 0 {
		return splits[size-1] + "." + suffix
	}

	return ""
}

func (atm *authMiddle) needAuth(c *ship.Context) error {
	c.SetRespHeader(ship.HeaderWWWAuthenticate, `Basic realm="Restricted"`)
	c.WriteHeader(http.StatusUnauthorized)

	return ship.ErrUnauthorized
}
