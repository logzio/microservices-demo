package payment

import (
	"github.com/go-kit/kit/log"
	"time"
	"github.com/go-kit/kit/log/level"

	//"os"
)

// LoggingMiddleware logs method calls, parameters, results, and elapsed time.
func LoggingMiddleware(logger log.Logger) Middleware {
	return func(next Service) Service {
		return loggingMiddleware{
			next:   next,
			logger: logger,
		}
	}
}

type loggingMiddleware struct {
	next   Service
	logger log.Logger
}

func (mw loggingMiddleware) Authorise(amount float32, traceID string) (auth Authorisation, err error) {
	defer func(begin time.Time) {
		if (err == nil) {
			level.Info(mw.logger).Log(
				"method", "Authorise",
				"result", auth.Authorised,
				"took", time.Since(begin),
				"traceID", traceID,
			)
		} else {
			level.Error(mw.logger).Log(
				"method", "Authorise",
				"result", auth.Authorised,
				"took", time.Since(begin),
				"traceID", traceID,
				"error", err,
			)
		}		
	}(time.Now())
	return mw.next.Authorise(amount, traceID)
}

func (mw loggingMiddleware) Health() (health []Health) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "Health",
			"result", len(health),
			"took", time.Since(begin),
		)
	}(time.Now())
	return mw.next.Health()
}
