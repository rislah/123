package local_test

import (
	"github.com/rislah/fakes/internal/local"
	"github.com/rislah/fakes/internal/tests"
	"testing"
)

func TestLocalDB(t *testing.T) {
	tests.TestUserDB(t, local.MakeUserDB)
}
