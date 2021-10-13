package throttler_test

import (
	"context"
	"github.com/alicebob/miniredis/v2"
	"github.com/cep21/circuit"
	"github.com/rislah/fakes/internal/redis"
	"github.com/rislah/fakes/internal/throttler"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type throttlerTestCase struct {
	th  throttler.Throttler
	rtl throttler.RateLimitTracker
	srv *miniredis.Miniredis

	maxTimeout         time.Duration
	baseTimeout        time.Duration
	attemptWindow      time.Duration
	attemptLimit       int64
	timeoutScaleFactor float64
	keyType            string
	ids                []throttler.ID
}

func TestThrottler(t *testing.T) {
	tests := []struct {
		name               string
		maxTimeout         time.Duration
		baseTimeout        time.Duration
		attemptWindow      time.Duration
		timeoutScaleFactor float64
		attemptLimit       int64
		keyType            string
		ids                []throttler.ID
		test               func(ctx context.Context, t *testing.T, testCase throttlerTestCase)
	}{
		{
			name:               "tryAttempt with no ids",
			maxTimeout:         0,
			baseTimeout:        0,
			attemptWindow:      0,
			timeoutScaleFactor: 0,
			keyType:            "test_key",
			attemptLimit:       0,
			ids:                []throttler.ID{},
			test: func(ctx context.Context, t *testing.T, testCase throttlerTestCase) {
				testCase.srv.FlushAll()
				shouldThrottle, err := testCase.th.TryAttempt(ctx, testCase.ids)
				assert.NoError(t, err)
				assert.False(t, shouldThrottle)
			},
		},
		{
			name:               "tryAttempt when num attempts less attempt limit",
			maxTimeout:         0,
			baseTimeout:        0,
			attemptWindow:      0,
			attemptLimit:       1,
			timeoutScaleFactor: 0,
			keyType:            "test_key",
			ids: []throttler.ID{
				{Key: "test_key", Type: "ip"},
			},
			test: func(ctx context.Context, t *testing.T, testCase throttlerTestCase) {
				testCase.srv.FlushAll()

				shouldThrottle, err := testCase.th.TryAttempt(ctx, testCase.ids)
				assert.NoError(t, err)
				assert.False(t, shouldThrottle)
			},
		},
		{
			name:               "tryAttempt should throttle when num attempts greater than attempt limit",
			maxTimeout:         1 * time.Minute,
			baseTimeout:        1 * time.Minute,
			attemptWindow:      5 * time.Minute,
			attemptLimit:       1,
			keyType:            "test_key1",
			timeoutScaleFactor: 0,
			ids:                []throttler.ID{{Key: "test_key1", Type: "ip"}},
			test: func(ctx context.Context, t *testing.T, testCase throttlerTestCase) {
				testCase.srv.FlushAll()

				incr := func(expected int64) {
					val, err := testCase.rtl.IncrAttempts(testCase.ids[0])
					assert.NoError(t, err)
					assert.Equal(t, expected, val)
				}

				incr(1)
				incr(2)

				shouldThrottle, err := testCase.th.TryAttempt(ctx, testCase.ids)
				assert.NoError(t, err)
				assert.True(t, shouldThrottle)
			},
		},
		{
			name:               "timedOut should return false if not timed out",
			maxTimeout:         1 * time.Minute,
			baseTimeout:        1 * time.Minute,
			attemptWindow:      5 * time.Minute,
			attemptLimit:       1,
			keyType:            "test_key2",
			timeoutScaleFactor: 0,
			ids:                []throttler.ID{{Key: "test_key2", Type: "ip"}},
			test: func(ctx context.Context, t *testing.T, testCase throttlerTestCase) {
				testCase.srv.FlushAll()

				timedOut, err := testCase.th.TimedOut(ctx, testCase.ids)
				assert.NoError(t, err)
				assert.False(t, timedOut)
			},
		},
		{
			name:               "timedOut should return true if timed out",
			maxTimeout:         1 * time.Minute,
			baseTimeout:        1 * time.Minute,
			attemptWindow:      5 * time.Minute,
			attemptLimit:       1,
			keyType:            "test_key3",
			timeoutScaleFactor: 0,
			ids:                []throttler.ID{{Key: "test_key3", Type: "ip"}},
			test: func(ctx context.Context, t *testing.T, testCase throttlerTestCase) {
				testCase.srv.FlushAll()

				incr := func(expected int64) {
					val, err := testCase.rtl.IncrAttempts(testCase.ids[0])
					assert.NoError(t, err)
					assert.Equal(t, expected, val)
				}

				incr(1)
				incr(2)

				shouldThrottle, err := testCase.th.TryAttempt(ctx, testCase.ids)
				assert.NoError(t, err)
				assert.True(t, shouldThrottle)

				timedOut, err := testCase.th.TimedOut(ctx, testCase.ids)
				assert.NoError(t, err)
				assert.True(t, timedOut)
			},
		},
		{
			name:               "incr should return throttle when timed out",
			maxTimeout:         1 * time.Minute,
			baseTimeout:        1 * time.Minute,
			attemptWindow:      5 * time.Minute,
			attemptLimit:       1,
			keyType:            "test_key4",
			timeoutScaleFactor: 0,
			ids:                []throttler.ID{{Key: "test_key4", Type: "ip"}},
			test: func(ctx context.Context, t *testing.T, testCase throttlerTestCase) {
				testCase.srv.FlushAll()

				incr := func(expected int64) {
					val, err := testCase.rtl.IncrAttempts(testCase.ids[0])
					assert.NoError(t, err)
					assert.Equal(t, expected, val)
				}

				incr(1)
				incr(2)

				shouldThrottle, err := testCase.th.TryAttempt(ctx, testCase.ids)
				assert.NoError(t, err)
				assert.True(t, shouldThrottle)

				shouldThrottleIncr, err := testCase.th.Incr(ctx, testCase.ids)
				assert.NoError(t, err)
				assert.True(t, shouldThrottleIncr)
			},
		},
		{
			name:               "incr should return when not timed out but past attempts limits",
			maxTimeout:         1 * time.Minute,
			baseTimeout:        1 * time.Minute,
			attemptWindow:      5 * time.Minute,
			attemptLimit:       1,
			keyType:            "test_key4",
			timeoutScaleFactor: 0,
			ids:                []throttler.ID{{Key: "test_key4", Type: "ip"}},
			test: func(ctx context.Context, t *testing.T, testCase throttlerTestCase) {
				testCase.srv.FlushAll()

				incr := func(expected int64) {
					val, err := testCase.rtl.IncrAttempts(testCase.ids[0])
					assert.NoError(t, err)
					assert.Equal(t, expected, val)
				}

				incr(1)
				incr(2)

				shouldThrottleIncr, err := testCase.th.Incr(ctx, testCase.ids)
				assert.NoError(t, err)
				assert.True(t, shouldThrottleIncr)
			},
		},
		{
			name:               "throttle accordingly to quota",
			maxTimeout:         24 * time.Hour,
			baseTimeout:        2 * time.Second,
			attemptWindow:      5 * time.Minute,
			timeoutScaleFactor: 1.0,
			attemptLimit:       1,
			keyType:            "test_key",
			ids:                []throttler.ID{{Key: "test_key", Type: "ip"}},
			test: func(ctx context.Context, t *testing.T, testCase throttlerTestCase) {
				testCase.srv.FlushAll()

				incr := func(expected int64) {
					val, err := testCase.rtl.IncrAttempts(testCase.ids[0])
					assert.NoError(t, err)
					assert.Equal(t, expected, val)
				}

				incr(1)
				incr(2)

				shouldThrottleIncr, err := testCase.th.Incr(ctx, testCase.ids)
				assert.NoError(t, err)
				assert.True(t, shouldThrottleIncr)

				<-time.After(2 * time.Second)

				shouldThrottleIncr, err = testCase.th.Incr(ctx, testCase.ids)
				assert.NoError(t, err)
				assert.False(t, shouldThrottleIncr)
			},
		},
	}

	for _, tc := range tests {
		test := tc
		t.Run(test.name, func(t *testing.T) {
			srv, err := miniredis.Run()
			assert.NoError(t, err)

			client, err := redis.NewClient(srv.Addr(), &circuit.Manager{}, nil)
			assert.NoError(t, err)

			config := throttler.Config{
				Client:             client,
				AttemptLimit:       test.attemptLimit,
				KeyType:            test.keyType,
				AttemptWindow:      test.attemptWindow,
				BaseTimeout:        test.baseTimeout,
				MaxTimeout:         test.maxTimeout,
				TimeoutScaleFactor: test.timeoutScaleFactor,
			}
			th := throttler.New(&config)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			test.test(ctx, t, throttlerTestCase{
				th:                 th,
				srv:                srv,
				rtl:                throttler.NewRateLimitTracker(config.Client, config.MaxTimeout, config.KeyType, config.AttemptWindow),
				attemptLimit:       config.AttemptLimit,
				baseTimeout:        config.BaseTimeout,
				maxTimeout:         config.MaxTimeout,
				timeoutScaleFactor: config.TimeoutScaleFactor,
				keyType:            test.keyType,
				ids:                test.ids,
			})
		})
	}
}
