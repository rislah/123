package queries_test

import (
	"testing"

	"github.com/rislah/fakes/internal/local"
	"github.com/rislah/fakes/internal/tests"
)

func TestLocalRoles(t *testing.T) {
	tests.TestRoles(t, local.MakeRoleDB, local.MakeUserDB)
}
