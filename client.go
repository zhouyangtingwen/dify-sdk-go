package dify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	host string
	apiSecretKey string

	httpClient *http.Client
	httpRequest *http.Request
}

func NewClientWithConfig(c *ClientConfig) *Client {
	var httpClient = &http.Client{}

	if c.Timeout == 0 {
		httpClient.Timeout = c.Timeout
	}
	if c.Transport != nil {
		httpClient.Transport = c.Transport
	}

	return &Client{
		host: c.Host,
		apiSecretKey: c.ApiSecretKey,
		httpClient: httpClient,
	}
}

func NewClient(host, apiSecretKey string) *Client {
	return NewClientWithConfig(&ClientConfig{
		Host: host,
		ApiSecretKey: apiSecretKey,
	})
}

func (c *Client) NewHttpRequest(ctx context.Context, method, requestUrl string, request ...interface{}) (r *http.Request, err error) {
	if method == http.MethodGet {
		if len(request) > 0 {
			if urlValues, ok := request[0].(url.Values); ok {
				var requestUrlParse *url.URL
				if requestUrlParse, err = url.Parse(requestUrl); err != nil {
					return
				}
				requestUrlParse.RawQuery = urlValues.Encode()
				requestUrl = requestUrlParse.String()
			}
		}
		r, err = http.NewRequestWithContext(ctx, method, requestUrl, http.NoBody)
	} else if method == http.MethodPost {
		var b io.Reader
		if len(request) > 0 {
			var reqBytes []byte
			if reqBytes, err = json.Marshal(request[0]); err != nil {
				return
			}
			b = bytes.NewBuffer(reqBytes)
		} else {
			b = http.NoBody
		}
		r, err = http.NewRequestWithContext(ctx, method, requestUrl, b)
	} else {
		err = errors.New("NewHttpRequest.method must be http.MethodGet or http.MethodPost")
	}
	return
}

func (c *Client) SetHttpRequest(r *http.Request) *Client {
	c.httpRequest = r
	return c
}

func (c *Client) SetHttpRequestHeader(key string, value string) *Client {
	c.httpRequest.Header.Set(key, value)
	return c
}

func (c *Client) SendRequest(res interface{}) (err error) {
	if c.httpRequest == nil {
		panic("http_request illegal")
	}
	var resp *http.Response
	if resp, err = c.httpClient.Do(c.httpRequest); err != nil {
		return
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(res)
	return
}

func (c *Client) SendRequestStream() (resp *http.Response, err error) {
	if c.httpRequest == nil {
		panic("http_request illegal")
	}
	resp, err = c.httpClient.Do(c.httpRequest)
	return
}

func (c *Client) GetHost() string {
	var host = strings.TrimSuffix(c.host, "/")
	return host
}

func (c *Client) GetApiSecretKey() string {
	return c.apiSecretKey
}

func (c *Client) Api() *Api {
	return &Api{
		c: c,
	}
}