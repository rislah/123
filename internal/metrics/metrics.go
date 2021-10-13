package metrics

import (
	"fmt"
	"github.com/uber-go/tally"
	"github.com/uber-go/tally/prometheus"
	"io"
	"time"
)

type Tags struct {
	Key   string
	Value string
}

type counter struct {
	name    string
	tags    Tags
	counter tally.Counter
}

type Metrics struct {
	counters []counter
	scope    tally.Scope
	closer   io.Closer
}

func New(reporter tally.CachedStatsReporter) Metrics {
	scope, closer := tally.NewRootScope(tally.ScopeOptions{
		Tags:           map[string]string{},
		Prefix:         "my_service",
		CachedReporter: reporter,
		Separator:      prometheus.DefaultSeparator,
	}, 1*time.Second)

	return Metrics{
		scope:    scope,
		closer:   closer,
		counters: []counter{},
	}
}

func (m *Metrics) NewCounter(name string, tags Tags) tally.Counter {
	c := m.scope.Tagged(map[string]string{tags.Key: tags.Value}).Counter(name)
	ctr := counter{
		name: name,
		tags:    tags,
		counter: c,
	}

	m.counters = append(m.counters, ctr)
	return c
}

func (m *Metrics) IncrementCounterWithTag(name string, tag Tags) {
	for _, v := range m.counters {
		if v.name == name && v.tags == tag {
			fmt.Println("jamh")
			v.counter.Inc(1)
		}
	}
}

func (m *Metrics) IncrementCounter(name string, tag Tags) {
	m.scope.Tagged(map[string]string{tag.Key: tag.Value}).Counter(name).Inc(1)
}