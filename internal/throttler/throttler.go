package throttler

import (
	"context"
	"fmt"
	"math"
	"time"

	errorx "github.com/rislah/fakes/internal/errors"
	"github.com/rislah/fakes/internal/redis"
)

type Throttler interface {
	TryAttempt(context.Context, []ID) (bool, error)
	TimedOut(context.Context, []ID) (bool, error)
	Reset(context.Context, []ID) error
	Incr(context.Context, []ID) (bool, error)
}

type Timeout struct {
	StartTime time.Time
	Duration  time.Duration
}

func (t *Timeout) EndTime() time.Time {
	return t.StartTime.Add(t.Duration)
}

func (t *Timeout) redisTimeoutStr() string {
	return fmt.Sprintf("%d:%d", t.StartTime.UnixNano(), int64(t.Duration))
}

type throttlerImpl struct {
	rlt                RateLimitTracker
	attemptLimit       int64
	baseTimeout        time.Duration
	maxTimeout         time.Duration
	timeoutScaleFactor float64
	shouldSkip         bool
}

type Header struct {
	Limit     string
	Remaining string
	ResetUnix string
}

type Config struct {
	Client             redis.Client
	KeyType            string
	AttemptLimit       int64
	AttemptWindow      time.Duration
	BaseTimeout        time.Duration
	MaxTimeout         time.Duration
	TimeoutScaleFactor float64
}

func New(c *Config) *throttlerImpl {
	return &throttlerImpl{
		rlt:                NewRateLimitTracker(c.Client, c.MaxTimeout, c.KeyType, c.AttemptWindow),
		attemptLimit:       c.AttemptLimit,
		baseTimeout:        c.BaseTimeout,
		maxTimeout:         c.MaxTimeout,
		timeoutScaleFactor: c.TimeoutScaleFactor,
		shouldSkip:         true,
	}
}

func (t *throttlerImpl) TryAttempt(ctx context.Context, ids []ID) (bool, error) {
	for _, uid := range ids {
		if uid.Key == "" {
			continue
		}

		numAttempts, err := t.rlt.IncrAttempts(uid)
		if err != nil {
			return false, errorx.Wrap(err, "TryAttempt")
		}

		if numAttempts > t.attemptLimit {
			if err := t.rlt.ResetAttempts(uid); err != nil {
				return true, errorx.Wrap(err, "TryAttempt")
			}

			lastTimeout, err := t.rlt.LastTimeout(uid)
			if err != nil {
				return true, errorx.Wrap(err, "TryAttempt")
			}

			var duration time.Duration
			if lastTimeout.Duration == 0 {
				duration = t.baseTimeout
			} else {
				duration = time.Duration(math.Min(float64(lastTimeout.Duration)*t.timeoutScaleFactor, float64(t.maxTimeout)))
			}

			now := time.Now()
			lastTimeout = Timeout{now, duration}
			if err := t.rlt.SetTimeout(uid, lastTimeout); err != nil {
				return true, errorx.Wrap(err, "TryAttempt")
			}

			return true, nil
		}
	}

	return false, nil
}

func (t *throttlerImpl) TimedOut(ctx context.Context, ids []ID) (bool, error) {
	for _, uid := range ids {
		timeout, err := t.rlt.LastTimeout(uid)
		if err != nil {
			return true, err
		}

		if time.Now().Before(timeout.EndTime()) {
			return true, nil
		}
	}

	return false, nil
}

func (t throttlerImpl) Reset(ctx context.Context, ids []ID) error {
	for _, uid := range ids {
		if err := t.rlt.ResetAttempts(uid); err != nil {
			return errorx.Wrap(err, "Reset")
		}

		if err := t.rlt.ResetTimeout(uid); err != nil {
			return errorx.Wrap(err, "Reset")
		}
	}

	return nil
}

func (t *throttlerImpl) Incr(ctx context.Context, ids []ID) (bool, error) {
	timedOut, err := t.TimedOut(ctx, ids)
	if err != nil {
		return true, err
	}

	if timedOut {
		return true, nil
	}

	attemptsExceeded, err := t.TryAttempt(ctx, ids)
	if err != nil {
		return true, err
	}

	return attemptsExceeded, nil
}
