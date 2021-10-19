package app_test

import (
	"testing"

	"github.com/rislah/fakes/internal/local"
	"github.com/rislah/fakes/internal/tests"
)

func TestLocalUserDB(t *testing.T) {
	tests.TestUserDB(t, local.MakeUserDB)
}
