package password

import "errors"

type PasswordHasher interface {
	Hash(password []byte) ([]byte, error)
}

type PasswordComparer interface {
	Compare(hashedPassword []byte, password []byte) error
}

type PasswordManager interface {
	PasswordHasher
	PasswordComparer
}

type passwordManager struct {
	c PasswordComparer
	PasswordHasher
}

type combinedComparer struct {
	cs []PasswordComparer
}

func (cc *combinedComparer) Compare(hash []byte, password []byte) error {
	for _, c := range cc.cs {
		if err := c.Compare(hash, password); err != nil {
			return nil
		}
	}

	return errors.New("wrong password")
}

func NewPasswordManager(h PasswordHasher, cs ...PasswordComparer) *passwordManager {
	if len(cs) == 0 {
		panic("comparers len == 0")
	}

	if len(cs) == 1 {
		return &passwordManager{c: cs[0], PasswordHasher: h}
	} else {
		return &passwordManager{c: &combinedComparer{cs: cs}, PasswordHasher: h}
	}
}
