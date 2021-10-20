package credentials_test

import (
	"github.com/rislah/fakes/internal/credentials"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPassword(t *testing.T) {
	tests := []struct {
		scenario string
		password credentials.Password
		test     func(p credentials.Password)
	}{
		{
			scenario: "should generate and verify bcrypt hash",
			password: "parool",
			test: func(p credentials.Password) {
				hash, err := p.GenerateBCrypt()
				assert.NoError(t, err)
				assert.NotEmpty(t, hash)

				compare, err := p.CompareBCrypt(hash)
				assert.NoError(t, err)
				assert.True(t, compare)
			},
		},
		{
			scenario: "should return false if wrong digest",
			password: "parool",
			test: func(p credentials.Password) {
				hash, err := credentials.NewPassword("asd").GenerateBCrypt()
				assert.NoError(t, err)
				compare, err := p.CompareBCrypt(hash)
				assert.NoError(t, err)
				assert.False(t, compare)
			},
		},
		{
			scenario: "non ascii",
			password: "ẀẀẀẀẀẀẀ",
			test: func(p credentials.Password) {
				_, err := p.Validate()
				assert.Error(t, err)
				assert.Equal(t, credentials.ErrPasswordNonASCII, err)
			},
		},
		{
			scenario: "min length",
			password: "1",
			test: func(p credentials.Password) {
				_, err := p.Validate()
				assert.Error(t, err)
				assert.Equal(t, credentials.ErrPasswordLength, err)
			},
		},
		{
			scenario: "max length",
			password: "Dqbq5Ci312rACp8jDLuWJuEnAEkYEZogjA8X5hVsza4CXDUZ0y9PYCi7kcNVP8JZgBLExAlaaa",
			test: func(p credentials.Password) {
				_, err := p.Validate()
				assert.Error(t, err)
				assert.Equal(t, credentials.ErrPasswordLength, err)
			},
		},
		{
			scenario: "complexity min",
			password: "123123123",
			test: func(p credentials.Password) {
				_, err := p.Validate()
				assert.Error(t, err)
				assert.Equal(t, credentials.ErrPasswordNotComplexEnough, err)
			},
		},
		{
			scenario: "complexity pass",
			password: "p@r00l!23",
			test: func(p credentials.Password) {
				_, err := p.Validate()
				assert.NoError(t, err)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			test.test(test.password)
		})
	}
}
