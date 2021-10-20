package app_test

import (
	"github.com/rislah/fakes/internal/local"
	"github.com/rislah/fakes/internal/tests"
	"testing"
)

func TestLocalAuthenticator(t *testing.T) {
	tests.TestAuthenticator(t, local.MakeUserDB)
}
