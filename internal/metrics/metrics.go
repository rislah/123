package metrics

import (
	"fmt"
	"time"

	"github.com/segmentio/stats/v4"
)

type Metrics interface {
    Inc(labelName string, tags ...stats.Tag)
    Measure(labelName string, t time.Duration, tags ...stats.Tag)
}

var (
	CircuitBreakerName = "circuitbreaker"
	HttpRequestName    = "http_requests"
)

type metricsImpl struct {
	engine *stats.Engine
}

func New(engine *stats.Engine) *metricsImpl {
	return &metricsImpl{engine}
}

func (m metricsImpl) Inc(labelName string, tags ...stats.Tag) {
	m.engine.Incr(fmt.Sprintf("%s_total", labelName), tags...)
}

func (m metricsImpl) Measure(labelName string, t time.Duration, tags ...stats.Tag) {
	m.engine.Observe(fmt.Sprintf("%s_time", labelName), t, tags...)
}
