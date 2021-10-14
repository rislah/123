package metrics

import (
	"fmt"
	"time"

	"github.com/segmentio/stats/v4"
)

var (
	OperationCountName = "operation_count"
	OperationTimeName  = "operation_time"
	CircuitBreakerName = "circuitbreaker"
	HttpRequestName    = "http_requests"
)

type Metrics struct {
	engine *stats.Engine
}

func New(engine *stats.Engine) Metrics {
	return Metrics{engine}
}

func (m Metrics) Inc(labelName string, tags ...stats.Tag) {
	m.engine.Incr(fmt.Sprintf("%s_total", labelName), tags...)
}

func (m Metrics) Measure(labelName string, t time.Duration, tags ...stats.Tag) {
	m.engine.Observe(fmt.Sprintf("%s_time", labelName), t, tags...)
}

func (m Metrics) IncCircuitBreaker(status string, tags ...stats.Tag) {
	m.engine.Incr(fmt.Sprintf("%s_%s", CircuitBreakerName, status), tags...)
}
