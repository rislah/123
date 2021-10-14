package tests

import (
	"testing"
	"time"

	"github.com/rislah/fakes/internal/redis"
	"github.com/rislah/fakes/internal/throttler"
	"github.com/stretchr/testify/assert"
)

type rateLimiterTestCase struct {
	redis            redis.Client
	rateLimitTracker throttler.RateLimitTracker

	ids []throttler.ID
}

type MakeRedis func() (redis.Client, error)

func TestRateLimitTracker(t *testing.T, makeRedis MakeRedis) {
	tests := []struct {
		name              string
		timeoutExpiration time.Duration
		attemptExpiration time.Duration
		ids               []throttler.ID
		test              func(t *testing.T, testCase rateLimiterTestCase)
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
			test: func(t *testing.T, testCase rateLimiterTestCase) {
				val, err := testCase.rateLimitTracker.IncrAttempts(testCase.ids[0])
				assert.NoError(t, err)
				assert.Equal(t, int64(1), val)

				val, err = testCase.rateLimitTracker.IncrAttempts(testCase.ids[0])
				assert.NoError(t, err)
				assert.Equal(t, int64(2), val)
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
			test: func(t *testing.T, testCase rateLimiterTestCase) {
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
			test: func(t *testing.T, testCase rateLimiterTestCase) {
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
			client, err := makeRedis()
			assert.NoError(t, err)

			key := tc.ids[0].Key
			tc := rateLimiterTestCase{
				rateLimitTracker: throttler.NewRateLimitTracker(client, tc.timeoutExpiration, key, tc.attemptExpiration),
				ids:              tc.ids,
				redis:            client,
			}

			test.test(t, tc)
		})
	}
}
