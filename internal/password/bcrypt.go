package password

import (
	"golang.org/x/crypto/bcrypt"
)

type bcryptPasswordHasher struct{}

func NewBcryptPasswordHasher() PasswordHasher {
	return &bcryptPasswordHasher{}
}

func (ph *bcryptPasswordHasher) Compare(hashedPassword []byte, password []byte) error {
	return bcrypt.CompareHashAndPassword(hashedPassword, password)
}

func (ph *bcryptPasswordHasher) Hash(password []byte) ([]byte, error) {
	return bcrypt.GenerateFromPassword(password, 12)
}
