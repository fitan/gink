package http

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
)

func ExamplePopulateRequestContext() {
	handler := NewServer(
		func(ctx context.Context, request interface{}) (response interface{}, err error) {
			fmt.Println("Method", ctx.Value(ContextKeyRequestMethod).(string))
			fmt.Println("RequestPath", ctx.Value(ContextKeyRequestPath).(string))
			fmt.Println("RequestURI", ctx.Value(ContextKeyRequestURI).(string))
			fmt.Println("X-Request-ID", ctx.Value(ContextKeyRequestXRequestID).(string))
			return struct{}{}, nil
		},
		func(context.Context, *gin.Context) (interface{}, error) { return struct{}{}, nil },
		func(context.Context, *gin.Context, interface{}) error { return nil },
		ServerBefore(PopulateRequestContext),
	)

	r := gin.New()
	r.PATCH("/search", handler.ServeHTTP)
	server := httptest.NewServer(r)
	defer server.Close()

	req, _ := http.NewRequest("PATCH", fmt.Sprintf("%s/search?q=sympatico", server.URL), nil)
	req.Header.Set("X-Request-Id", "a1b2c3d4e5")
	http.DefaultClient.Do(req)

	// Output:
	// Method PATCH
	// RequestPath /search
	// RequestURI /search?q=sympatico
	// X-Request-ID a1b2c3d4e5
}
