package throttler_test

import (
	"testing"
	"time"

	"github.com/rislah/fakes/internal/local"
	"github.com/rislah/fakes/internal/redis"
	"github.com/rislah/fakes/internal/throttler"
	"github.com/stretchr/testify/assert"
)

func TestRateLimitTracker(t *testing.T) {
	tests := []struct {
		name              string
		timeoutExpiration time.Duration
		attemptExpiration time.Duration
		id                throttler.ID
		test              func(t *testing.T, id throttler.ID, timeoutExpiration, attemptExpiration time.Duration, rds redis.Client)
	}{
		{
			name: "increment",
			id: throttler.ID{
				Key:  "test_key1",
				Type: "ip",
			},
			timeoutExpiration: 1 * time.Minute,
			attemptExpiration: 1 * time.Minute,
			test: func(t *testing.T, id throttler.ID, timeoutExpiration, attemptExpiration time.Duration, rds redis.Client) {
				assert.NoError(t, rds.FlushAll())

				thr := throttler.NewRateLimitTracker(rds, timeoutExpiration, id.Key, attemptExpiration)
				val, err := thr.IncrAttempts(id)
				assert.NoError(t, err)
				assert.Equal(t, int64(1), val)

				keys, err := rds.Keys("*")
				assert.NoError(t, err)
				assert.Len(t, keys, 1)
			},
		},
		{
			name: "reset attempt",
			id: throttler.ID{
				Key:  "test_key",
				Type: "ip",
			},
			timeoutExpiration: 1 * time.Minute,
			attemptExpiration: 1 * time.Minute,
			test: func(t *testing.T, id throttler.ID, timeoutExpiration, attemptExpiration time.Duration, rds redis.Client) {
				assert.NoError(t, rds.FlushAll())

				thr := throttler.NewRateLimitTracker(rds, timeoutExpiration, id.Key, attemptExpiration)
				val, err := thr.IncrAttempts(id)
				assert.NoError(t, err)
				assert.Equal(t, int64(1), val)

				err = thr.ResetAttempts(id)
				assert.NoError(t, err)

				keys, err := rds.Keys("*")
				assert.NoError(t, err)
				assert.Len(t, keys, 0)
			},
		},
		{
			name: "set timeout and reset",
			id: throttler.ID{
				Key:  "test_key",
				Type: "ip",
			},
			timeoutExpiration: 1 * time.Minute,
			attemptExpiration: 1 * time.Minute,
			test: func(t *testing.T, id throttler.ID, timeoutExpiration, attemptExpiration time.Duration, rds redis.Client) {
				assert.NoError(t, rds.FlushAll())
				thr := throttler.NewRateLimitTracker(rds, timeoutExpiration, id.Key, attemptExpiration)
				err := thr.SetTimeout(id, throttler.Timeout{StartTime: time.Now(), Duration: 1 * time.Minute})
				assert.NoError(t, err)

				keys, err := rds.Keys("*")
				assert.NoError(t, err)
				assert.Len(t, keys, 1)

				err = thr.ResetTimeout(id)
				assert.NoError(t, err)

				keys, err = rds.Keys("*")
				assert.NoError(t, err)
				assert.Len(t, keys, 0)
			},
		},
	}

	for _, tc := range tests {
		test := tc
		t.Run(tc.name, func(t *testing.T) {
			client, err := local.NewRedis()
			assert.NoError(t, err)
			test.test(t, tc.id, tc.timeoutExpiration, tc.attemptExpiration, client)
		})
	}
}
