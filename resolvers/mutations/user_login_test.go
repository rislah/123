package mutations_test

import (
	"testing"

	"github.com/rislah/fakes/internal/local"
	"github.com/rislah/fakes/internal/tests"
)

func TestLocalLogin(t *testing.T) {
	tests.TestLogin(t, local.MakeRoleDB, local.MakeUserDB)
}
