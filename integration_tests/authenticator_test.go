package integration_tests

import (
	"testing"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rislah/fakes/internal/tests"
)

func TestIntegrationAuthenticator(t *testing.T) {
	tests.TestAuthenticator(t, makeUserDB)
}
