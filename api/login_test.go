package api_test

import (
	"testing"

	"github.com/rislah/fakes/internal/local"
	"github.com/rislah/fakes/internal/tests"
)

func TestLocalLogin(t *testing.T) {
	tests.TestAPILogin(t, local.MakeUserDB, local.MakeNoopMetrics, local.MakeRedis)
}
