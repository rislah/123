package metrics

import (
	"time"

	"github.com/cep21/circuit/v3"
	"github.com/segmentio/stats/v4"
)

type CommandFactory struct {
	Metrics Metrics
}

func (c *CommandFactory) CommandProperties(name string) circuit.Config {
	return circuit.Config{
		Metrics: circuit.MetricsCollectors{
			Run: []circuit.RunMetrics{
				&runMetricsCollector{
					name,
					c.Metrics,
				},
			},
		},
	}
}

type runMetricsCollector struct {
	name    string
	metrics Metrics
}

func (cm runMetricsCollector) ErrShortCircuit(now time.Time) {
	cm.metrics.Inc(CircuitBreakerName, stats.T("name", cm.name), stats.T("error", "short_circuit"))
}

func (cm runMetricsCollector) ErrConcurrencyLimitReject(now time.Time) {
	cm.metrics.Inc(CircuitBreakerName, stats.T("name", cm.name), stats.T("error", "concurrency_limit"))
}

func (cm runMetricsCollector) ErrFailure(now time.Time, duration time.Duration) {
	cm.metrics.Inc(CircuitBreakerName, stats.T("name", cm.name), stats.T("error", "failure"))
}

func (cm runMetricsCollector) ErrTimeout(now time.Time, duration time.Duration) {
	cm.metrics.Inc(CircuitBreakerName, stats.T("name", cm.name), stats.T("error", "timeout"))
}
func (cm runMetricsCollector) Success(now time.Time, duration time.Duration) {
}

func (cm runMetricsCollector) ErrBadRequest(now time.Time, duration time.Duration) {
}

func (cm runMetricsCollector) ErrInterrupt(now time.Time, duration time.Duration) {
}
