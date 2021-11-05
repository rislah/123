package app_test

import (
	"testing"

	"github.com/rislah/fakes/internal/local"
	"github.com/rislah/fakes/internal/tests"
)

func TestLocalAuthenticator(t *testing.T) {
	tests.TestAuthenticator(t, local.MakeRoleDB, local.MakeUserDB)
}
