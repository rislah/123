package integration_tests

import (
	"testing"

	"github.com/rislah/fakes/internal/tests"
)

func TestIntegrationUserDB(t *testing.T) {
	tests.TestUserDB(t, makeUserDB, makeRoleDB)
}
