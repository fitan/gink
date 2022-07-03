package http

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/lb"
	"github.com/go-kit/log"
	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"io"
	"net/url"
	"os"
)

func ipFactory() sd.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			return instance, nil
		}, nil, nil
	}
}

type BeforeKitLBConf struct {
	balancerFunc func(s sd.Endpointer) lb.Balancer
}

type BeforeKitLbOption func(conf *BeforeKitLBConf)

func WithBeforeKitLbRandom(seed int64) BeforeKitLbOption {
	return func(conf *BeforeKitLBConf) {
		conf.balancerFunc = func(e sd.Endpointer) lb.Balancer {
			return lb.NewRandom(e, seed)
		}
	}
}

func BeforeKitLb(instance sd.Instancer, seed int64) resty.RequestMiddleware {
	e := sd.NewEndpointer(instance, ipFactory(), log.NewLogfmtLogger(os.Stdout))

	var balance lb.Balancer

	if seed == 0 {
		balance = lb.NewRoundRobin(e)
	} else {
		balance = lb.NewRandom(e, seed)
	}

	return func(c *resty.Client, req *resty.Request) error {
		be, err := balance.Endpoint()
		if err != nil {
			err = errors.Wrap(err, "endpoint error")
			return err
		}
		ip, _ := be(req.Context(), nil)

		urlParse, err := url.Parse(req.URL)
		if err != nil {
			err = errors.Wrap(err, "url parse error")
			return err
		}

		urlParse.Host = ip.(string)
		urlParse.Scheme = "http"

		req.URL = urlParse.String()

		return nil
	}
}
