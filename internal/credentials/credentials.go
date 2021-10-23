package credentials

import (
	"unicode"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rislah/fakes/internal/errors"
)

var (
	authenticationFailureCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "authentication_failure_total",
	})
)

func init() {
	prometheus.Register(authenticationFailureCounter)
}

const (
	bcryptCost        = 10
	passwordMinLength = 8
	passwordMaxLength = 70
	minZxcvbnScore    = 2
)

type Credentials struct {
	Username Username
	Password Password
}

func New(username string, password string) Credentials {
	return Credentials{
		Username: NewUsername(username),
		Password: NewPassword(password),
	}
}

func (c Credentials) Valid() error {
	if c.Username.String() == "" {
		return ErrUsernameMissing
	}

	if c.Password.String() == "" {
		return ErrPasswordMissing
	}

	if err := c.Password.ValidateLength(); err != nil {
		return err
	}

	if err := c.Username.ValidateLength(); err != nil {
		return err
	}

	if err := c.Username.ValidateRegex(); err != nil {
		return err
	}

	return nil
}

func ComparePassword(digest string, pass Password) error {
	compare, err := pass.CompareBCrypt(digest)
	if err != nil {
		return errors.New(err)
	}

	if !compare {
		authenticationFailureCounter.Inc()
		return ErrPasswordMismatch
	}

	return nil
}

func isASCII(str string) bool {
	for _, c := range str {
		if c > unicode.MaxASCII {
			return false
		}
	}

	return true
}
