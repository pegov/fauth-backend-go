package password

import (
	"bytes"
	"errors"
)

type plainTextPasswordHasher struct{}

func NewPlainTextPasswordHasher() PasswordManager {
	return &plainTextPasswordHasher{}
}

func (ph *plainTextPasswordHasher) Compare(
	hashedPassword []byte,
	password []byte,
) error {
	if bytes.Equal(hashedPassword, password) {
		return nil
	} else {
		return errors.New("password mismatch")
	}
}

func (ph *plainTextPasswordHasher) Hash(password []byte) ([]byte, error) {
	return password, nil
}
