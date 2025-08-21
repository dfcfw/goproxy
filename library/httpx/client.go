package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func NewClient(clis ...*http.Client) Client {
	cli := Client{cli: http.DefaultClient}
	if len(clis) != 0 {
		cli.cli = clis[0]
	}

	return cli
}

// Client http 客户端。
type Client struct {
	cli *http.Client
}

// JSON 发送 GET 请求，响应 JSON 数据。
func (c Client) JSON(ctx context.Context, rawURL string, header http.Header, result any) error {
	resp, err := c.sendJSON(ctx, http.MethodGet, rawURL, header, http.NoBody)
	if err != nil {
		return err
	}
	err = c.unmarshalJSON(resp.Body, result)

	return err
}

// PostJSON 通过 POST 发送 JSON 报文，响应 JSON 数据。
func (c Client) PostJSON(ctx context.Context, rawURL string, header http.Header, body, result any) error {
	resp, err := c.sendJSON(ctx, http.MethodPost, rawURL, header, body)
	if err != nil {
		return err
	}
	err = c.unmarshalJSON(resp.Body, result)

	return err
}

// PostForm POST FromData 数据，注意只是普通的 FormData，不支持发送二进制文件。
func (c Client) PostForm(ctx context.Context, rawURL string, header http.Header, body url.Values, result any) error {
	if header == nil {
		header = make(http.Header, 4)
	}
	header.Set("Content-Type", "application/x-www-form-urlencoded")

	encode := body.Encode()
	req, err := c.newRequest(ctx, http.MethodPost, rawURL, header, strings.NewReader(encode))
	if err != nil {
		return err
	}

	resp, err := c.do(req)
	if err != nil {
		return err
	}

	return c.unmarshalJSON(resp.Body, result)
}

// RoundTrip 发送请求。
func (c Client) RoundTrip(req *http.Request) (*http.Response, error) {
	return c.do(req)
}

func (c Client) sendJSON(ctx context.Context, method, rawURL string, header http.Header, body any) (*http.Response, error) {
	if header == nil {
		header = make(http.Header, 4)
	}
	header.Set("Accept", "application/json")

	var r io.Reader
	if method != http.MethodGet && method != http.MethodHead {
		rd, err := c.marshalJSON(body)
		if err != nil {
			return nil, err
		}
		r = rd
		header.Set("Content-Type", "application/json; charset=utf-8")
	}

	req, err := c.newRequest(ctx, method, rawURL, header, r)
	if err != nil {
		return nil, err
	}

	return c.do(req)
}

func (Client) newRequest(ctx context.Context, method, rawURL string, header http.Header, body io.Reader) (*http.Request, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	req, err := http.NewRequestWithContext(ctx, method, rawURL, body)
	if err != nil {
		return nil, err
	}
	if len(header) != 0 {
		req.Header = header
	}

	return req, nil
}

func (c Client) do(req *http.Request) (*http.Response, error) {
	h := req.Header
	if host := h.Get("Host"); host != "" {
		req.Host = host
	}
	if h.Get("User-Agent") == "" {
		chrome133 := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36"
		h.Set("User-Agent", chrome133)
	}
	if h.Get("Accept-Language") == "" {
		h.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")
	}

	const retry = 3
	const delay = 500 * time.Millisecond
	res, err := c.send(req, retry, delay)
	if err != nil {
		return nil, err
	}
	code := res.StatusCode
	if code >= 200 && code < 400 {
		return res, nil
	}

	//goland:noinspection GoUnhandledErrorResult
	defer res.Body.Close()

	e := &Error{
		Code:    code,
		Request: req,
	}
	buf := make([]byte, 1024)
	n, _ := res.Body.Read(buf)
	e.Body = buf[:n]

	return nil, e
}

// send 如果请求失败重试几次，以及重试间隔。
func (c Client) send(req *http.Request, retry int, delay time.Duration) (*http.Response, error) {
	cli := c.getClient()
	res, err := cli.Do(req)
	if retry > 0 && c.needRetry(err, res) {
		retry--
		if res != nil { // 关闭上一次的 resp
			_ = res.Body.Close()
		}

		ctx := req.Context()
		c.sleep(ctx, delay)

		// 需要注意的是，req 能否被重复发送，取决于多种情况，
		// 比如：req.Body 没有被消费等，使用时一定要注意。
		res, err = cli.Do(req)
	}

	return res, err
}

func (Client) needRetry(err error, res *http.Response) bool {
	if err != nil {
		return true
	}
	code := res.StatusCode

	return code == http.StatusTooManyRequests ||
		(code >= 500 && code < 600)
}

func (Client) sleep(ctx context.Context, d time.Duration) {
	if d <= 0 {
		return
	}

	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}

func (c Client) marshalJSON(v any) (io.Reader, error) {
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(v)
	return buf, err
}

func (c Client) getClient() *http.Client {
	if cli := c.cli; cli != nil {
		return cli
	}
	return http.DefaultClient
}

func (c Client) unmarshalJSON(rc io.ReadCloser, v any) error {
	//goland:noinspection GoUnhandledErrorResult
	defer rc.Close()
	if v == nil || rc == http.NoBody {
		return nil
	}

	return json.NewDecoder(rc).Decode(v)
}
