package casauth

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
)

type StringURL string

func (s StringURL) Configure(context.Context) (string, error) {
	return string(s), nil
}

type Configurer interface {
	Configure(ctx context.Context) (string, error)
}

type Client interface {
	Auth(ctx context.Context, name, passwd string) error
}

func NewClient(cfg Configurer, rtp http.RoundTripper, log *slog.Logger) Client {
	return casClient{
		cfg: cfg,
		rtp: rtp,
		log: log,
	}
}

type casClient struct {
	cfg Configurer
	rtp http.RoundTripper
	log *slog.Logger
}

func (c casClient) Auth(ctx context.Context, name, passwd string) error {
	attrs := []any{slog.String("name", name)}
	c.log.DebugContext(ctx, "开始CAS认证", attrs...)
	strURL, err := c.cfg.Configure(ctx)
	if err != nil {
		attrs = append(attrs, slog.Any("error", err))
		c.log.ErrorContext(ctx, "获取CAS配置错误", attrs...)
		return err
	}

	reqURL, err := url.Parse(strURL)
	if err != nil {
		attrs = append(attrs, slog.Any("error", err))
		c.log.ErrorContext(ctx, "获取CAS服务器地址错误", attrs...)
		return err
	}

	sum := md5.Sum([]byte(passwd))
	pwd := hex.EncodeToString(sum[:])
	query := reqURL.Query()
	query.Set("usrNme", name)
	query.Set("passwd", pwd)
	if query.Get("devTyp") == "" {
		query.Set("devTyp", "pc")
	}
	reqURL.RawQuery = query.Encode()
	destURL := reqURL.String()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, destURL, nil)
	if err != nil {
		attrs = append(attrs, slog.Any("error", err))
		c.log.ErrorContext(ctx, "构造 http.Request 错误", attrs...)
		return err
	}
	resp, err := c.rtp.RoundTrip(req)
	if err != nil {
		attrs = append(attrs, slog.Any("error", err))
		c.log.ErrorContext(ctx, "请求CAS服务器错误", attrs...)
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	res := new(responseBody)
	if err = json.NewDecoder(resp.Body).Decode(res); err != nil {
		attrs = append(attrs, slog.Any("error", err))
		c.log.ErrorContext(ctx, "读取 CAS 响应数据错误", attrs...)
		return err
	}

	return res.checkError()
}

var errorCodes = map[string]string{
	"01": "密码错误",
	"03": "用户不存在",
	"04": "设备类型错误",
	"09": "用户名或密码为空",
}

// reply sso 认证服务的响应报文
type responseBody struct {
	RspCde string `json:"rspCde"` // 业务响应码
	RspMsg string `json:"rspMsg"` // 响应消息
}

func (rb responseBody) checkError() error {
	code := rb.RspCde
	if code == "00" {
		return nil
	}

	msg := rb.RspMsg
	if msg == "" {
		msg = errorCodes[code]
	}
	if msg == "" {
		msg = "认证错误（cas-server-code: " + code + "）"
	}

	return errors.New(msg)
}
