package circuitbreaker

import (
	"time"

	"github.com/cep21/circuit"
	"github.com/cep21/circuit/closers/hystrix"
)

func NewDefault() *circuit.Manager {
	openerFactory := func(circuitName string) hystrix.ConfigureOpener {
		return hystrix.ConfigureOpener{
			ErrorThresholdPercentage: 50,
			RequestVolumeThreshold:   10,
			RollingDuration:          10 * time.Second,
			NumBuckets:               10,
		}
	}

	closerFactory := func(circuitName string) hystrix.ConfigureCloser {
		return hystrix.ConfigureCloser{
			SleepWindow:                  5 * time.Second,
			HalfOpenAttempts:             1,
			RequiredConcurrentSuccessful: 1,
		}
	}

	hystrixFactory := hystrix.Factory{
		CreateConfigureOpener: []func(circuitName string) hystrix.ConfigureOpener{openerFactory},
		CreateConfigureCloser: []func(circuitName string) hystrix.ConfigureCloser{closerFactory},
	}

	manager := &circuit.Manager{
		DefaultCircuitProperties: []circuit.CommandPropertiesConstructor{
			hystrixFactory.Configure,
		},
	}

	return manager
}
