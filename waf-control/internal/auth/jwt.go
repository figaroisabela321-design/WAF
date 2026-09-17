package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenService issues and validates JWTs.
type TokenService interface {
	Issue(claims *Claims, expireHours int) (token string, expiresIn int, err error)
	Parse(token string) (*Claims, error)
}

type JWTService struct {
	secret []byte
}

func NewJWTService(secret string) *JWTService {
	return &JWTService{secret: []byte(secret)}
}

type jwtClaims struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

func (s *JWTService) Issue(c *Claims, expireHours int) (string, int, error) {
	if expireHours <= 0 {
		expireHours = 24
	}
	expiresIn := expireHours * 3600
	now := time.Now()
	claims := jwtClaims{
		UserID: c.UserID, Username: c.Username, Roles: c.Roles, Permissions: c.Permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(expireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   c.UserID,
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := t.SignedString(s.secret)
	if err != nil {
		return "", 0, err
	}
	return signed, expiresIn, nil
}

func (s *JWTService) Parse(tokenStr string) (*Claims, error) {
	t, err := jwt.ParseWithClaims(tokenStr, &jwtClaims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}
	jc, ok := t.Claims.(*jwtClaims)
	if !ok || !t.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return &Claims{
		UserID: jc.UserID, Username: jc.Username, Roles: jc.Roles, Permissions: jc.Permissions,
	}, nil
}
