package circuitbreaker

import (
	"time"

	"github.com/cep21/circuit/v3"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	circuitManagerCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "circuitbreaker_total",
	}, []string{"name", "status"})

	circuitManagerDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "circuitbreaker_duration_seconds",
	}, []string{"name", "status"})
)

func init() {
	prometheus.Register(circuitManagerCounter)
	prometheus.Register(circuitManagerDuration)
}

type commandFactory struct {
}

func NewMetricsCommandFactory() commandFactory {
	return commandFactory{}
}

func (c commandFactory) CommandProperties(name string) circuit.Config {
	return circuit.Config{
		Metrics: circuit.MetricsCollectors{
			Run: []circuit.RunMetrics{
				&runMetricsCollector{
					name,
				},
			},
		},
	}
}

type runMetricsCollector struct {
	name string
}

func (cm runMetricsCollector) ErrShortCircuit(now time.Time) {
	circuitManagerCounter.WithLabelValues(cm.name, "short_circuit").Inc()
}

func (cm runMetricsCollector) ErrConcurrencyLimitReject(now time.Time) {
	circuitManagerCounter.WithLabelValues(cm.name, "concurrency_limit").Inc()
}

func (cm runMetricsCollector) ErrFailure(now time.Time, duration time.Duration) {
	circuitManagerCounter.WithLabelValues(cm.name, "failure").Inc()
	circuitManagerDuration.WithLabelValues(cm.name, "failure").Observe(duration.Seconds())
}

func (cm runMetricsCollector) ErrTimeout(now time.Time, duration time.Duration) {
	circuitManagerCounter.WithLabelValues(cm.name, "timeout").Inc()
	circuitManagerDuration.WithLabelValues(cm.name, "timeout").Observe(duration.Seconds())
}

func (cm runMetricsCollector) ErrBadRequest(now time.Time, duration time.Duration) {
	circuitManagerCounter.WithLabelValues(cm.name, "bad_request").Inc()
	circuitManagerDuration.WithLabelValues(cm.name, "bad_request").Observe(duration.Seconds())
}

func (cm runMetricsCollector) ErrInterrupt(now time.Time, duration time.Duration) {
	circuitManagerCounter.WithLabelValues(cm.name, "interrupt").Inc()
	circuitManagerDuration.WithLabelValues(cm.name, "interrupt").Observe(duration.Seconds())
}

func (cm runMetricsCollector) Success(now time.Time, duration time.Duration) {
	circuitManagerCounter.WithLabelValues(cm.name, "success").Inc()
	circuitManagerDuration.WithLabelValues(cm.name, "success").Observe(duration.Seconds())
}
