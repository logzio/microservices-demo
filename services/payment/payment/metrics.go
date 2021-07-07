package payment

import (
	"github.com/go-kit/kit/metrics"
	"time"
	"fmt"
)


type InstrumentingMiddleware struct {
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
	next           Service
}

func (mw InstrumentingMiddleware) Authorise(amount float32, traceID string) (auth Authorisation, err error) {
	defer func(begin time.Time) {
		lvs := []string{"route", "paymentAuth","method", "POST", "error", fmt.Sprint(err != nil), "isWS", "false", "status_code", "200"}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.Authorise(amount, traceID)
}

func (mw InstrumentingMiddleware) Health() (health []Health) {
	defer func(begin time.Time) {
		lvs := []string{"route", "health","method", "GET", "error", "false", "isWS", "false", "status_code", "200"}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.Health()
}