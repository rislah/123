package app_test

import (
	"testing"

	"github.com/rislah/fakes/internal/local"
	"github.com/rislah/fakes/internal/tests"
)

func TestLocalRoleDB(t *testing.T) {
	tests.TestRoleDB(t, local.MakeRoleDB)
}
