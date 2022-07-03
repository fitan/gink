package http

import (
	"github.com/go-resty/resty/v2"
)

func WithReqHeader(header string, value string) ReqOption {
	return func(r *resty.Request) {
		r.SetHeader(header, value)
	}
}

func WithReqHeaders(headers map[string]string) ReqOption {
	return func(r *resty.Request) {
		r.SetHeaders(headers)
	}
}

func WithReqPathParam(param string, value string) ReqOption {
	return func(r *resty.Request) {
		r.SetPathParam(param, value)
	}
}

func WithReqPathParams(params map[string]string) ReqOption {
	return func(r *resty.Request) {
		r.SetPathParams(params)
	}
}

func WithReqQueryParam(param string, value string) ReqOption {
	return func(r *resty.Request) {
		r.SetQueryParam(param, value)
	}
}

func WithReqQueryParams(params map[string]string) ReqOption {
	return func(r *resty.Request) {
		r.SetQueryParams(params)
	}
}
