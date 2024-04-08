package password

type PasswordHasher interface {
	Compare(hashedPassword []byte, password []byte) error
	Hash(password []byte) ([]byte, error)
}
