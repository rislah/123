package throttler_test

import (
	"github.com/alicebob/miniredis/v2"
	"github.com/cep21/circuit"
	"github.com/rislah/fakes/internal/redis"
	"github.com/rislah/fakes/internal/throttler"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type testCase struct {
	rateLimitTracker throttler.RateLimitTracker
	redis            redis.Client
	ids              []throttler.ID
}

func TestRateLimitTracker(t *testing.T) {
	tests := []struct {
		name              string
		timeoutExpiration time.Duration
		attemptExpiration time.Duration
		ids               []throttler.ID
		test              func(t *testing.T, testCase testCase)
	}{
		{
			name: "increment two times",
			ids: []throttler.ID{
				{
					Key:  "test_key1",
					Type: "ip",
				},
			},
			timeoutExpiration: 1 * time.Minute,
			attemptExpiration: 1 * time.Minute,
			test: func(t *testing.T, testCase testCase) {
				val, err := testCase.rateLimitTracker.IncrAttempts(testCase.ids[0])
				assert.NoError(t, err)
				assert.Equal(t, int64(1), val)

				val, err = testCase.rateLimitTracker.IncrAttempts(testCase.ids[0])
				assert.NoError(t, err)
				assert.Equal(t, int64(2), val)
			},
		},
		{
			name:              "reset attempts success",
			timeoutExpiration: 1 * time.Minute,
			attemptExpiration: 1 * time.Minute,
			ids: []throttler.ID{
				{
					Key:  "test_key2",
					Type: "ip",
				},
			},
			test: func(t *testing.T, testCase testCase) {
				_, err := testCase.rateLimitTracker.IncrAttempts(testCase.ids[0])
				assert.NoError(t, err)

				v, err := testCase.rateLimitTracker.ResetAttempts(testCase.ids[0])
				assert.NoError(t, err)

				assert.Equal(t, true, v)
			},
		},
		{
			name:              "last timeout empty",
			timeoutExpiration: 1 * time.Minute,
			attemptExpiration: 1 * time.Minute,
			ids: []throttler.ID{
				{
					Key:  "test_key3",
					Type: "ip",
				},
			},
			test: func(t *testing.T, testCase testCase) {
				timeout, err := testCase.rateLimitTracker.LastTimeout(testCase.ids[0])
				assert.NoError(t, err)
				assert.Equal(t, time.Duration(0), timeout.Duration)
			},
		},
		{
			name:              "last timeout exists",
			timeoutExpiration: 1 * time.Minute,
			attemptExpiration: 1 * time.Minute,
			ids: []throttler.ID{
				{
					Key:  "test_key4",
					Type: "ip",
				},
			},
			test: func(t *testing.T, testCase testCase) {
				duration := 1 * time.Second

				err := testCase.rateLimitTracker.SetTimeout(testCase.ids[0], throttler.Timeout{
					StartTime: time.Now(),
					Duration:  duration,
				})
				assert.NoError(t, err)

				timeout, err := testCase.rateLimitTracker.LastTimeout(testCase.ids[0])
				assert.NoError(t, err)
				assert.Equal(t, duration, timeout.Duration)
			},
		},
	}

	for _, tc := range tests {
		test := tc
		t.Run(tc.name, func(t *testing.T) {
			srv, err := miniredis.Run()
			assert.NoError(t, err)

			client, err := redis.NewClient(srv.Addr(), &circuit.Manager{}, nil)
			assert.NoError(t, err)

			key := tc.ids[0].Key
			tc := testCase{
				rateLimitTracker: throttler.NewRateLimitTracker(client, tc.timeoutExpiration, key, tc.attemptExpiration),
				ids:              tc.ids,
				redis:            client,
			}

			test.test(t, tc)
		})
	}
}
