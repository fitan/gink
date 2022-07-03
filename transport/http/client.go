package http

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
)

// Client wraps a URL and provides a method that implements endpoint.Endpoint.
type Client struct {
	client *resty.Client
}

func (c *Client) Endpoint(url ReqOption, enc RestyEncodeRequestFunc, dec RestyDecodeResponseFunc, options ...RequestOption) endpoint.Endpoint {
	r := c.client.R()
	url(r)
	request := &Request{
		req:            makeCreateRequestFunc(r, enc),
		dec:            dec,
		before:         make([]RestyRequestFunc, 0),
		after:          make([]RestyResponseFunc, 0),
		finalizer:      make([]ClientFinalizerFunc, 0),
		bufferedStream: false,
	}
	for _, option := range options {
		option(request)
	}

	return request.Endpoint()

}

type Request struct {
	req            RestyCreateRequestFunc
	dec            RestyDecodeResponseFunc
	before         []RestyRequestFunc
	after          []RestyResponseFunc
	finalizer      []ClientFinalizerFunc
	bufferedStream bool
}

func (r *Request) Endpoint() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		ctx, cancel := context.WithCancel(ctx)

		var (
			resp *resty.Response
			err  error
		)
		if r.finalizer != nil {
			defer func() {
				if resp != nil {
					ctx = context.WithValue(ctx, ContextKeyResponseHeaders, resp.Header)
					ctx = context.WithValue(ctx, ContextKeyResponseSize, resp.Size())
				}
				for _, f := range r.finalizer {
					f(ctx, err)
				}
			}()
		}

		req, err := r.req(ctx, request)
		if err != nil {
			cancel()
			return nil, err
		}

		for _, f := range r.before {
			ctx = f(ctx, req)
		}

		resp, err = req.SetContext(ctx).Send()

		if err != nil {
			cancel()
			return nil, err
		}

		if r.bufferedStream {
			resp.RawResponse.Body = bodyWithCancel{ReadCloser: resp.RawResponse.Body, cancel: cancel}
		} else {
			defer resp.RawResponse.Body.Close()
			defer cancel()
		}

		for _, f := range r.after {
			ctx = f(ctx, resp)
		}

		response, err := r.dec(ctx, resp)
		if err != nil {
			return nil, err
		}

		return response, nil
	}
}

// NewClient constructs a usable Client for a single remote method.
func NewClient(client *resty.Client, options ...ClientOption) *Client {
	c := &Client{
		client: client,
	}
	for _, option := range options {
		option(c)
	}

	return c
}

// ClientOption sets an optional parameter for clients.
type ClientOption func(client *Client)

type RequestOption func(request *Request)

type ReqOption func(request *resty.Request)

func SetClientDebug(b bool) ClientOption {
	return func(client *Client) {
		client.client = client.client.SetDebug(b)
	}
}

func Req(method string, url string, options ...ReqOption) ReqOption {
	return func(r *resty.Request) {
		r.URL = url
		r.Method = method
		for _, option := range options {
			option(r)
		}
	}
}

// SetClient sets the underlying HTTP client used for requests.
// By default, http.DefaultClient is used.
func SetClient(client *resty.Client) ClientOption {
	return func(c *Client) { c.client = client }
}

type bodyWithCancel struct {
	io.ReadCloser

	cancel context.CancelFunc
}

func (bwc bodyWithCancel) Close() error {
	bwc.ReadCloser.Close()
	bwc.cancel()
	return nil
}

// ClientFinalizerFunc can be used to perform work at the end of a client HTTP
// request, after the response is returned. The principal
// intended use is for error logging. Additional response parameters are
// provided in the context under keys with the ContextKeyResponse prefix.
// Note: err may be nil. There maybe also no additional response parameters
// depending on when an error occurs.
type ClientFinalizerFunc func(ctx context.Context, err error)

func EncodeJSONRequest(c context.Context, r *resty.Request, request interface{}) error {
	r.SetBody(request)
	r.SetContext(c)
	return nil
}

func DecodeJSONResponse(i interface{}) RestyDecodeResponseFunc {
	return func(ctx context.Context, resp *resty.Response) (interface{}, error) {
		if resp.StatusCode() != http.StatusOK {
			return resp.String(), fmt.Errorf("unexpected status code %d", resp.StatusCode())
		}

		if is, ok := i.(*string); ok {
			*is = resp.String()
			return resp, nil
		}

		result := gjson.GetBytes(resp.Body(), "code")
		if !result.Exists() {
			err := json.Unmarshal(resp.Body(), i)
			if err != nil {
				err = errors.Wrap(err, "unmarshal response")
				return resp.String(), err
			}
			return resp.String(), nil
		}

		if result.Int() != http.StatusOK {
			s := gjson.Get(resp.String(), "err").String()
			return resp.String(), fmt.Errorf("response err: %s", s)
		}

		dataResult := gjson.GetBytes(resp.Body(), "data")
		err := json.Unmarshal(resp.Body()[dataResult.Index:result.Index+len(result.Raw)], i)
		if err != nil {
			err = errors.Wrap(err, "unmarshal response data")
			return resp.String(), err
		}

		return resp, nil
	}
}

func makeCreateRequestFunc(req *resty.Request, enc RestyEncodeRequestFunc) RestyCreateRequestFunc {
	return func(ctx context.Context, i interface{}) (*resty.Request, error) {
		err := enc(ctx, req, i)
		if err != nil {
			return nil, err
		}
		return req, nil
	}
}
