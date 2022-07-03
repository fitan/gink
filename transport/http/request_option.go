package http

import (
	"context"
	"encoding/json"
	"github.com/go-resty/resty/v2"
	"os"
)

func WithRequestDebug() RequestOption {
	return func(request *Request) {
		request.before = append(request.before, func(ctx context.Context, req *resty.Request) context.Context {
			req.EnableTrace()
			return ctx
		})
		request.after = append(request.after, func(ctx context.Context, resp *resty.Response) context.Context {
			info := DebugInfo{
				Response: SetResponse(resp),
				Request:  SetRequest(resp.Request),
				Info:     SetInfo(resp.Request.TraceInfo()),
			}
			_ = json.NewEncoder(os.Stdout).Encode(info)
			return ctx
		})
	}
}
