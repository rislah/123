package throttler

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	pkgRedis "github.com/go-redis/redis/v8"
	errorx "github.com/rislah/fakes/internal/errors"
	"github.com/rislah/fakes/internal/redis"
)

var (
	ErrThrottlerUnableToIncrAttempt = &errorx.WrappedError{
		StatusCode: http.StatusInternalServerError,
		CodeValue:  "throttler_unable_to_increment_attempt",
		Message:    "unable to increment attempt",
		ShouldLog:  true,
	}
	ErrThrottlerUnableToGetLastTimeout = &errorx.WrappedError{
		StatusCode: http.StatusInternalServerError,
		CodeValue:  "throttler_unable_to_get_last_timeout",
		Message:    "unable to get last timeout",
		ShouldLog:  true,
	}
	ErrThrottlerUnableToSetTimeout = &errorx.WrappedError{
		StatusCode: http.StatusInternalServerError,
		CodeValue:  "throttler_unable_to_set_timeout",
		Message:    "unable to set timeout",
		ShouldLog:  true,
	}
	ErrThrottlerUnableToResetAttempts = &errorx.WrappedError{
		StatusCode: http.StatusInternalServerError,
		CodeValue:  "throttler_unable_to_reset_attempts",
		Message:    "unable to reset attempts",
		ShouldLog:  true,
	}
	ErrThrottlerUnableToResetTimeout = &errorx.WrappedError{
		StatusCode: http.StatusInternalServerError,
		CodeValue:  "throttler_unable_to_reset_timeout",
		Message:    "unable to reset timeout",
		ShouldLog:  true,
	}
)

type Throttler interface {
	TryAttempt(context.Context, []ID) (bool, error)
	TimedOut(context.Context, []ID) (bool, error)
	Reset(context.Context, []ID) error
	Incr(context.Context, []ID) (bool, error)
	Details(context.Context, []ID) *Header
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
			if _, err := t.rlt.ResetAttempts(uid); err != nil {
				return true, ErrThrottlerUnableToResetAttempts
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
		if _, err := t.rlt.ResetAttempts(uid); err != nil {
			return errorx.Wrap(err, "Reset")
		}

		if err := t.rlt.ResetTimeout(uid); err != nil {
			return errorx.Wrap(err, "Reset")
		}
	}

	return nil
}

func (t *throttlerImpl) Incr(ctx context.Context, ids []ID) (bool, error) {
	// if t.shouldSkip {
	// 	return false, nil
	// }

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

func (t *throttlerImpl) Details(ctx context.Context, ids []ID) *Header {
	var expiresIn int64
	var remaining int

	timedOut, err := t.TimedOut(ctx, ids)
	if err != nil {
		log.Fatalln(err)
	}

	duration := func(id ID, kind Kind) (int64, error) {
		duration, err := t.rlt.ExpiresAt(id, TimeoutKind)
		if err != nil {
			return -1, err
		}

		expiresIn = time.Now().Add(duration).Unix()
		return expiresIn, nil
	}

	switch timedOut {
	case true:
		expiresIn, err = duration(ids[0], TimeoutKind)
	case false:
		expiresIn, err = duration(ids[0], AttemptsKind)
	}

	if err != nil {
		if err == pkgRedis.Nil {
			expiresIn = 0
		} else {
			log.Fatalln(err)
		}
	}

	attempts, err := t.rlt.GetAttempts(ids[0])
	if err != nil {
		if err == pkgRedis.Nil {
			attempts = 0
		} else {
			log.Fatalln(err)
		}
	}

	remaining = int(t.attemptLimit - attempts)

	return &Header{
		Limit:     strconv.Itoa(int(t.attemptLimit)),
		Remaining: strconv.Itoa(remaining),
		ResetUnix: strconv.Itoa(int(expiresIn)),
	}
}
