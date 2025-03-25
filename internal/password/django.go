package password

import (
	"bytes"
	"crypto/pbkdf2"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
)

type DjangoPasswordHasher struct{}

func NewDjangoPasswordHasher() PasswordHasher {
	return &DjangoPasswordHasher{}
}

func (ph *DjangoPasswordHasher) Compare(hashedPassword []byte, password []byte) error {
	hasher := strings.SplitN(string(hashedPassword), "$", 2)[0]

	switch hasher {
	case "pbkdf2_sha256":
		if checkPBKDF2SHA256(string(hashedPassword), string(password)) {
			return nil
		} else {
			return errors.New("wrong password")
		}
	}

	return errors.New("UNIMPLEMENTED")
}

func (ph *DjangoPasswordHasher) Hash(_ []byte) ([]byte, error) {
	return nil, errors.New("UNIMPLEMENTED")
}

func checkPBKDF2SHA256(hashedPassword, password string) bool {
	parts := strings.SplitN(hashedPassword, "$", 4)
	iter, _ := strconv.Atoi(parts[1])
	salt := []byte(parts[2])
	k, _ := base64.StdEncoding.DecodeString(parts[3])
	key, _ := pbkdf2.Key(sha256.New, password, salt, iter, sha256.Size)
	return bytes.Equal(k, key)
}
