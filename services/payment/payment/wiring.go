package payment

import (
	"net/http"
	"os"

	"github.com/go-kit/kit/log"
	"golang.org/x/net/context"

	stdprometheus "github.com/prometheus/client_golang/prometheus"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
)

var (
	requestLatency = kitprometheus.NewHistogramFrom(stdprometheus.HistogramOpts{
		Name:    "request_duration_seconds",
		Help:    "Time (in seconds) spent serving HTTP requests.",
		Buckets: stdprometheus.DefBuckets,
	}, []string{"method", "route", "status_code", "isWS", "error"})
	requestCount = kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Name:      "request_count",
		Help:      "Number of requests received.",
	}, []string{"method", "route", "status_code", "isWS", "error"})
)

func WireUp(ctx context.Context, declineAmount float32, serviceName string) (http.Handler, log.Logger) {
	// Log domain.
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
	}

	// Service domain.
	var service Service
	{
		service = NewAuthorisationService(declineAmount)
		service = LoggingMiddleware(logger)(service)
		service = InstrumentingMiddleware{requestCount,requestLatency,service}
	}

	// Endpoint domain.
	endpoints := MakeEndpoints(service)

	router := MakeHTTPHandler(ctx, endpoints, logger)

	return router, logger
}
