package http

import (
	"github.com/go-resty/resty/v2"
	"net"
	"time"
)

type DebugInfo struct {
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
	DNSLookup      string   `json:"dns_lookup"`
	ConnTime       string   `json:"conn_time"`
	TCPConnTime    string   `json:"tcp_conn_time"`
	TLSHandshake   string   `json:"tls_handshake"`
	ServerTime     string   `json:"server_time"`
	ResponseTime   string   `json:"response_time"`
	TotalTime      string   `json:"total_time"`
	IsConnReused   bool     `json:"is_conn_reused"`
	IsConnWasIdle  bool     `json:"is_conn_was_idle"`
	ConnIdleTime   string   `json:"conn_idle_time"`
	RequestAttempt int      `json:"request_attempt"`
	RemoteAddr     net.Addr `json:"remote_addr"`
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
		DNSLookup:      traceInfo.DNSLookup.Truncate(time.Millisecond).String(),
		ConnTime:       traceInfo.ConnTime.Truncate(time.Millisecond).String(),
		TCPConnTime:    traceInfo.TCPConnTime.Truncate(time.Millisecond).String(),
		TLSHandshake:   traceInfo.TLSHandshake.Truncate(time.Millisecond).String(),
		ServerTime:     traceInfo.ServerTime.Truncate(time.Millisecond).String(),
		ResponseTime:   traceInfo.ResponseTime.Truncate(time.Millisecond).String(),
		TotalTime:      traceInfo.TotalTime.Truncate(time.Millisecond).String(),
		IsConnReused:   traceInfo.IsConnReused,
		IsConnWasIdle:  traceInfo.IsConnWasIdle,
		ConnIdleTime:   traceInfo.ConnIdleTime.Truncate(time.Millisecond).String(),
		RequestAttempt: traceInfo.RequestAttempt,
		RemoteAddr:     traceInfo.RemoteAddr,
	}
}
