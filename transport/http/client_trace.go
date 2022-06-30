package http

import (
	"context"
	"encoding/json"
	"github.com/go-resty/resty/v2"
	"net"
	"os"
	"time"
)

type TraceInfo struct {
	Response ResponseInfo `json:"response,omitempty"`

	Request RequestInfo `json:"request,omitempty"`

	Info Info `json:"trace_info,omitempty"`
}

type RequestInfo struct {
	Url        string      `json:"url"`
	Method     string      `json:"method"`
	QueryParam string      `json:"query_param"`
	Body       interface{} `json:"body"`
}

type ResponseInfo struct {
	StatusCode int         `json:"status_code"`
	Body       interface{} `json:"body"`
	Proto      string      `json:"proto"`
	ReceivedAt string      `json:"received_at"`
}

type Info struct {
	DNSLookup      time.Duration `json:"dns_lookup"`
	ConnTime       time.Duration `json:"conn_time"`
	TCPConnTime    time.Duration `json:"tcp_conn_time"`
	TLSHandshake   time.Duration `json:"tls_handshake"`
	ServerTime     time.Duration `json:"server_time"`
	ResponseTime   time.Duration `json:"response_time"`
	TotalTime      time.Duration `json:"total_time"`
	IsConnReused   bool          `json:"is_conn_reused"`
	IsConnWasIdle  bool          `json:"is_conn_was_idle"`
	ConnIdleTime   time.Duration `json:"conn_idle_time"`
	RequestAttempt int           `json:"request_attempt"`
	RemoteAddr     net.Addr      `json:"remote_addr"`
}

func SetResponse(resp *resty.Response) ResponseInfo {
	data := ResponseInfo{}
	data.StatusCode = resp.StatusCode()
	data.Proto = resp.Proto()
	data.ReceivedAt = resp.ReceivedAt().String()
	data.Body = resp.Body()
	return data
}

func SetRequest(req *resty.Request) RequestInfo {
	data := RequestInfo{}
	data.Body = req.Body
	data.Url = req.URL
	data.Method = req.Method
	data.QueryParam = req.QueryParam.Encode()
	return data
}

func SetInfo(traceInfo resty.TraceInfo) Info {
	return Info{
		DNSLookup:      traceInfo.DNSLookup,
		ConnTime:       traceInfo.ConnTime,
		TCPConnTime:    traceInfo.TCPConnTime,
		TLSHandshake:   traceInfo.TLSHandshake,
		ServerTime:     traceInfo.ServerTime,
		ResponseTime:   traceInfo.ResponseTime,
		TotalTime:      traceInfo.TotalTime,
		IsConnReused:   traceInfo.IsConnReused,
		IsConnWasIdle:  traceInfo.IsConnWasIdle,
		ConnIdleTime:   traceInfo.ConnIdleTime,
		RequestAttempt: traceInfo.RequestAttempt,
		RemoteAddr:     traceInfo.RemoteAddr,
	}
}

func WithTrace() RequestOption {
	return func(request *Request) {
		request.before = append(request.before, func(ctx context.Context, req *resty.Request) context.Context {
			req.EnableTrace()
			return ctx
		})
		request.after = append(request.after, func(ctx context.Context, resp *resty.Response) context.Context {
			info := TraceInfo{
				Response: SetResponse(resp),
				Request:  SetRequest(resp.Request),
				Info:     SetInfo(resp.Request.TraceInfo()),
			}
			_ = json.NewEncoder(os.Stdout).Encode(info)
			return ctx
		})
	}
}
