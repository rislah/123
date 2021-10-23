package api_test

import (
	"testing"

	"github.com/rislah/fakes/internal/local"
	"github.com/rislah/fakes/internal/tests"
)

func TestLocalGetUsers(t *testing.T) {
	tests.TestAPIGetUsers(t, local.MakeUserDB, local.MakeRedis)
}
