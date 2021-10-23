package circuitbreaker

import (
	"time"

	"github.com/cep21/circuit"
	"github.com/cep21/circuit/closers/hystrix"
)

var defaultConfig = Config{
	MaxConcurrentRequests:  5000,
	Timeout:                100 * time.Millisecond,
	RequestVolumeThreshold: 50,
	ErrorPercentThreshold:  50,
	SleepWindow:            500 * time.Millisecond,
}

type Config struct {
	MaxConcurrentRequests  int
	Timeout                time.Duration
	RequestVolumeThreshold int
	ErrorPercentThreshold  int
	SleepWindow            time.Duration
}

func mergeConfig(defaultConfig, newConfig Config) Config {
	c := newConfig
	if c.ErrorPercentThreshold == 0 {
		c.ErrorPercentThreshold = defaultConfig.ErrorPercentThreshold
	}

	if c.MaxConcurrentRequests == 0 {
		c.MaxConcurrentRequests = defaultConfig.MaxConcurrentRequests
	}

	if c.RequestVolumeThreshold == 0 {
		c.RequestVolumeThreshold = defaultConfig.RequestVolumeThreshold
	}

	if c.SleepWindow == 0 {
		c.SleepWindow = defaultConfig.SleepWindow
	}

	if c.Timeout == 0 {
		c.Timeout = defaultConfig.Timeout
	}

	return c
}

func New(name string, conf Config) (*circuit.Circuit, error) {
	mergedConfig := mergeConfig(defaultConfig, conf)

	openerFactory := func(circuitName string) hystrix.ConfigureOpener {
		return hystrix.ConfigureOpener{
			ErrorThresholdPercentage: int64(mergedConfig.ErrorPercentThreshold),
			RequestVolumeThreshold:   int64(mergedConfig.RequestVolumeThreshold),
		}
	}

	closerFactory := func(circuitName string) hystrix.ConfigureCloser {
		return hystrix.ConfigureCloser{
			SleepWindow: defaultConfig.SleepWindow,
		}
	}

	hystrixFactory := hystrix.Factory{
		CreateConfigureOpener: []func(circuitName string) hystrix.ConfigureOpener{openerFactory},
		CreateConfigureCloser: []func(circuitName string) hystrix.ConfigureCloser{closerFactory},
	}

	configureExecution := circuit.Config{
		Execution: circuit.ExecutionConfig{
			MaxConcurrentRequests: int64(mergedConfig.MaxConcurrentRequests),
			Timeout:               defaultConfig.Timeout,
		},
	}

	manager := &circuit.Manager{
		DefaultCircuitProperties: []circuit.CommandPropertiesConstructor{
			hystrixFactory.Configure,
			NewMetricsCommandFactory().CommandProperties,
		},
	}

	return manager.CreateCircuit(name, configureExecution)
}
