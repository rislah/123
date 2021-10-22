package credentials

import (
	"fmt"
	"github.com/nbutton23/zxcvbn-go"
	"github.com/rislah/fakes/internal/errors"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)


var (
	ErrPasswordNotComplexEnough = &errors.WrappedError{
		Code: http.StatusBadRequest,
		Msg:  "Password is not complex enough.",
	}

	ErrPasswordLength = &errors.WrappedError{
		Code: http.StatusBadRequest,
		Msg:  fmt.Sprintf("Password must be between %d and %d characters", passwordMinLength, passwordMaxLength),
	}

	ErrPasswordNonASCII = &errors.WrappedError{
		Code: http.StatusBadRequest,
		Msg:  "Password must consist of ASCII characters",
	}

	ErrPasswordMismatch = &errors.WrappedError{
		Code: http.StatusBadRequest,
		Msg:  "Incorrect username or password",
	}

	ErrPasswordMissing = &errors.WrappedError{
		Code: http.StatusBadRequest,
		Msg:  "Password is required",
	}
)

type Password string

func (p Password) String() string {
	return string(p)
}

func NewPassword(password string) Password {
	return Password(password)
}

func (p Password) GenerateBCrypt() (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(p), bcryptCost)
	if err != nil {
		return "", errors.Wrap(err, "failed to generate bcrypt hash")
	}

	return string(b), nil
}

func (p Password) CompareBCrypt(digest string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(digest), []byte(p))
	switch err {
	case nil:
		return true, nil
	case bcrypt.ErrMismatchedHashAndPassword:
		return false, nil
	default:
		return false, err
	}
}

func (p Password) ValidateLength() error {
	if len(p) < passwordMinLength || len(p) > passwordMaxLength {
		return ErrPasswordLength
	}
	return nil
}

func (p Password) ValidateStrength(userInputs ...string) (int, error) {
	//if !isASCII(string(p)) {
	//	return -1, ErrPasswordNonASCII
	//}

	//if err := p.ValidateLength(); err != nil {
	//	return -1, err
	//}

	strength := zxcvbn.PasswordStrength(string(p), userInputs)
	if strength.Score < minZxcvbnScore {
		return strength.Score, ErrPasswordNotComplexEnough
	}

	return strength.Score, nil
}
