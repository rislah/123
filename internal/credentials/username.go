package credentials

import (
	"fmt"
	"github.com/rislah/fakes/internal/errors"
	"net/http"
	"regexp"
)

var (
	validLoginRegex = regexp.MustCompile("^[a-z0-9][a-z0-9_]*$")

	ErrUsernameMissing = &errors.WrappedError{
		Code: http.StatusBadRequest,
		Msg:  "Username is required",
	}

	ErrUsernameRegexFail = &errors.WrappedError{
		Code: http.StatusBadRequest,
		Msg:  "Username must not have symbols, non-ASCII or uppercase characters",
	}

	ErrUsernameLength = &errors.WrappedError{
		Code: http.StatusBadRequest,
		Msg:  fmt.Sprintf("Username must be between %d and %d characters", usernameMinLength, usernameMaxLength),
	}

	usernameMinLength = 4
	usernameMaxLength = 120
)

type Username string

func NewUsername(username string) Username {
	return Username(username)
}

func (u Username) String() string {
	return string(u)
}

func (u Username) ValidateRegex() error {
	if !validLoginRegex.MatchString(u.String()) {
		return ErrUsernameRegexFail
	}
	
	return nil
}

func (u Username) ValidateLength() error {
	if len(u) > usernameMaxLength || len(u) < usernameMinLength {
		return ErrUsernameLength
	}

	return nil
}
