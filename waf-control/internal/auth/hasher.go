package auth

import "golang.org/x/crypto/bcrypt"

// PasswordHasher hashes and verifies passwords.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}

// BcryptHasher implements PasswordHasher with bcrypt.
type BcryptHasher struct {
	cost int
}

func NewBcryptHasher() *BcryptHasher {
	return &BcryptHasher{cost: bcrypt.DefaultCost}
}

func (h *BcryptHasher) Hash(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (h *BcryptHasher) Compare(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
