package throttler

import (
	"fmt"
	"github.com/rislah/fakes/internal/errors"
	"github.com/rislah/fakes/internal/redis"
	"strconv"
	"strings"
	"time"
)

type ID struct {
	Key  string
	Type string
}

type Kind string

const (
	AttemptsKind Kind = "attempts"
	TimeoutKind  Kind = "timeout"
)

type RateLimitTracker interface {
	IncrAttempts(uid ID) (int64, error)
	GetAttempts(uid ID) (int64, error)
	ResetAttempts(uid ID) (bool, error)
	LastTimeout(uid ID) (Timeout, error)
	SetTimeout(uid ID, timeout Timeout) error
	ResetTimeout(uid ID) error
	ExpiresAt(uid ID, kind Kind) (time.Duration, error)
}

type rateLimitTrackerImpl struct {
	redis             redis.Client
	keyType           string
	timeoutExpiration time.Duration
	attemptExpiration string
}

func NewRateLimitTracker(client redis.Client, timeoutExpiration time.Duration, keyType string, attemptExpiration time.Duration) *rateLimitTrackerImpl {
	return &rateLimitTrackerImpl{
		redis:             client,
		keyType:           keyType,
		timeoutExpiration: timeoutExpiration,
		attemptExpiration: strconv.FormatInt(int64(attemptExpiration/time.Second), 10),
	}
}

func (r *rateLimitTrackerImpl) attemptsKey(uid ID) string {
	return fmt.Sprintf("%s/%s/{%s}/attempts", r.keyType, uid.Type, uid.Key)
}

func (r *rateLimitTrackerImpl) timeoutKey(uid ID) string {
	return fmt.Sprintf("%s/%s/{%s}/timeout", r.keyType, uid.Type, uid.Key)
}

var incExpire = `
local current
current = redis.call("incr", KEYS[1])
redis.call("expire", KEYS[1], tonumber(ARGV[1]))
return tonumber(current)
`

func (r *rateLimitTrackerImpl) IncrAttempts(uid ID) (int64, error) {
	key := r.attemptsKey(uid)
	numHolder, err := r.redis.Eval(incExpire, []string{key}, []string{r.attemptExpiration})
	if err != nil {
		return -1, err
	}

	return numHolder.(int64), nil
}

func (r rateLimitTrackerImpl) ResetAttempts(uid ID) (bool, error) {
	return r.redis.Del(r.attemptsKey(uid))
}

func (r rateLimitTrackerImpl) LastTimeout(uid ID) (Timeout, error) {
	timeoutStr, err := r.redis.Get(r.timeoutKey(uid))
	if err != nil {
		if errors.IsWrappedRedisNilError(err) {
			return Timeout{}, nil
		}

		return Timeout{}, err
	}

	return timeoutFromRedisTimeoutStr(timeoutStr)
}

func (r rateLimitTrackerImpl) SetTimeout(uid ID, timeout Timeout) error {
	return r.redis.Set(r.timeoutKey(uid), timeout.redisTimeoutStr(), r.timeoutExpiration)
}

func (r rateLimitTrackerImpl) ResetTimeout(uid ID) error {
	_, err := r.redis.Del(r.timeoutKey(uid))
	return err
}

func timeoutFromRedisTimeoutStr(redisStr string) (Timeout, error) {
	timeoutSplit := strings.Split(redisStr, ":")

	startUnixNano, err := strconv.ParseInt(timeoutSplit[0], 10, 64)
	if err != nil {
		return Timeout{}, err
	}
	start := time.Unix(0, startUnixNano)

	durationNano, err := strconv.ParseInt(timeoutSplit[1], 10, 64)
	if err != nil {
		return Timeout{}, err
	}

	return Timeout{start, time.Duration(durationNano)}, nil
}

func (r *rateLimitTrackerImpl) ExpiresAt(uid ID, kind Kind) (time.Duration, error) {
	var url string

	switch kind {
	case AttemptsKind:
		url = r.attemptsKey(uid)
	case TimeoutKind:
		url = r.timeoutKey(uid)
	}

	return r.redis.TTL(url)
}

func (r *rateLimitTrackerImpl) GetAttempts(uid ID) (int64, error) {
	return r.redis.GetInt64(r.timeoutKey(uid))
}