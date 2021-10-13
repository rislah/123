package app_test

import (
	app "github.com/rislah/fakes/internal"
	"github.com/rislah/fakes/internal/local"
	"github.com/rislah/fakes/internal/tests"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUserService(t *testing.T) {
	tests.TestService(t, local.MakeUserDB)
}

func TestMissingDatabase(t *testing.T) {
	t.Parallel()
	defer func() {
		r := recover()
		assert.NotNil(t, r)
	}()
	app.NewUserService(app.ServiceConfig{UserDB: nil})
}
