package jwt_test

import (
	"testing"

	"github.com/rislah/fakes/internal/jwt"
	"github.com/stretchr/testify/assert"
)

func TestJWT(t *testing.T) {
	tests := []struct {
		scenario string
		test     func()
	}{
		{
			scenario: "encode and decode",
			test: func() {
				wrapper := jwt.NewHS256Wrapper("secret")
				claims := jwt.NewUserClaims("user", "guest")
				token, err := wrapper.Encode(claims)
				assert.NoError(t, err)
				assert.NotEmpty(t, token)

				decodedToken, err := wrapper.Decode(token, &jwt.UserClaims{})
				assert.NoError(t, err)

				uc, ok := decodedToken.Claims.(*jwt.UserClaims)
				assert.True(t, ok)
				assert.Equal(t, uc.Username, claims.Username)
				assert.Equal(t, uc.Role, claims.Role)
			},
		},
	}

	for _, test := range tests {
		t.Run(t.Name(), func(t *testing.T) {
			test.test()
		})
	}
}
