package local

import (
	"time"

	"github.com/rislah/fakes/internal/metrics"
	"github.com/segmentio/stats/v4"
)

type noopMetricsImpl struct {}

func NewNoopMetrics() *noopMetricsImpl {
    return &noopMetricsImpl{}
}

func MakeNoopMetrics() metrics.Metrics {
    return &noopMetricsImpl{}
}

func (n noopMetricsImpl) Inc(labelname string, tags ...stats.Tag) {

}

func (n noopMetricsImpl) Measure(labelname string, t time.Duration, tags ...stats.Tag) {
}
