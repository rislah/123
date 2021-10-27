package mutations_test

import (
	"testing"

	"github.com/rislah/fakes/internal/local"
	"github.com/rislah/fakes/internal/tests"
)

func TestLocalRegister(t *testing.T) {
	tests.TestRegister(t, local.MakeUserDB, local.MakeRoleDB)
}
