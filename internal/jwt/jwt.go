package jwt

import (
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/rislah/fakes/internal/errors"
)

const (
	expiresIn = 24 * time.Hour
)

var (
	ErrJWTAlgMismatch = &errors.WrappedError{
		Code: errors.ErrBadRequest,
		Msg:  "JWT algorithm mismatch",
	}
	ErrJWTInvalid = &errors.WrappedError{
		Code: errors.ErrBadRequest,
		Msg:  "Invalid JWT provided",
	}
)

type UserClaims struct {
	*jwt.RegisteredClaims
	Username string `json:"username"`
	Role     string `json:"role"`
}

func NewRegisteredClaims(expiresIn time.Duration) jwt.RegisteredClaims {
	now := time.Now()
	return jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
		IssuedAt:  jwt.NewNumericDate(now),
	}
}

func NewUserClaims(username, role string) UserClaims {
	rc := NewRegisteredClaims(expiresIn)
	uc := UserClaims{
		RegisteredClaims: &rc,
		Username:         username,
		Role:             role,
	}
	return uc
}

type Wrapper struct {
	Algorithm jwt.SigningMethod
	Secret    string
}

func NewHS256Wrapper(secret string) Wrapper {
	return Wrapper{
		Algorithm: jwt.SigningMethodHS256,
		Secret:    secret,
	}
}

func (w Wrapper) Encode(claims jwt.Claims) (string, error) {
	switch w.Algorithm {
	case jwt.SigningMethodHS256:
		token := jwt.NewWithClaims(w.Algorithm, claims)
		return token.SignedString([]byte(w.Secret))
	default:
		return "", errors.New("unknown JWT signing algorithm")
	}
}

func (w Wrapper) Decode(tokenStr string, claims jwt.Claims) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(w.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, ErrJWTInvalid
	}

	if token.Method.Alg() != w.Algorithm.Alg() {
		return nil, ErrJWTAlgMismatch
	}

	return token, nil
}
