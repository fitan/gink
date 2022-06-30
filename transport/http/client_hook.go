package http

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"strconv"
	"time"
)

func BeforeDebug(log *zap.SugaredLogger, bodyLimit int) resty.RequestMiddleware {
	return func(c *resty.Client, req *resty.Request) error {
		debug, ok := req.Context().Value(ContextKeyRequestDebug).(bool)
		if ok && debug {
			c.EnableTrace()

			var body string
			lenValue := req.Header.Get("Content-Length")
			if lenValue == "" {
				body = "header Content-Length unknown"
			}
			lenValueInt, err := strconv.Atoi(lenValue)
			if err != nil {
				body = fmt.Sprintf("header Content-Length strconv err: %s", err.Error())
			}
			if lenValueInt > bodyLimit {
				body = fmt.Sprintf("***** REQUEST TOO LARGE (size - %d) *****", lenValueInt)
			}

			_log := log.With(
				zap.String(
					"traceId", trace.SpanFromContext(req.Context()).SpanContext().TraceID().String(),
				),
			)

			_log.Infow(
				"resty response debug",
				req.Method, req.RawRequest.URL.RequestURI(),
				"HOST", req.RawRequest.URL.Host,
				"BODY", body,
			)
		}
		return nil
	}
}

func AfterDebug(log *zap.SugaredLogger, bodyLimit int64) resty.ResponseMiddleware {
	return func(c *resty.Client, res *resty.Response) error {
		debug, ok := res.Request.Context().Value(ContextKeyRequestDebug).(bool)
		if ok && debug {
			var body string
			if res.Size() > bodyLimit {
				body = fmt.Sprintf("***** RESPONSE TOO LARGE (size - %d) *****", res.Size())
			}
			_log := log.With(zap.String("traceId", trace.SpanFromContext(res.Request.Context()).SpanContext().TraceID().String()))

			_log.Infow("resty response debug",
				"STATUS", res.Status(),
				"PROTO", res.RawResponse.Proto,
				"RECEIVED AT", res.ReceivedAt().Format(time.RFC3339Nano),
				"TIME DURATION", res.Time(),
				"BODY", body,
			)

			info := res.Request.TraceInfo()

			_log.Infow("resty trace",
				"DNSLookup", info.DNSLookup,
				"ConnTime", info.ConnTime,
				"TCPConnTime", info.TCPConnTime,
				"TLSHandshake", info.TLSHandshake,
				"ServerTime", info.ServerTime,
				"ResponseTime", info.ResponseTime,
				"TotalTime", info.TotalTime,
				"IsConnReused", info.IsConnReused,
				"IsConnWasIdle", info.IsConnWasIdle,
				"ConnIdleTime", info.ConnIdleTime,
				"RequestAttempt", info.RequestAttempt,
				"RemoteAddr", info.RemoteAddr.String(),
			)
		}

		return nil
	}
}
