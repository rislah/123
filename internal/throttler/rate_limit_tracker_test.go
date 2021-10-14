package throttler_test

import (
	"testing"

	"github.com/rislah/fakes/internal/local"
	"github.com/rislah/fakes/internal/tests"
)

func TestRateLimitTracker(t *testing.T) {
	tests.TestRateLimitTracker(t, local.MakeRedis)
}
