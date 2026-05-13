package auth

import (
	"errors"
	"time"

	"github.com/alloy/diffinder/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenType string

const (
	TokenAccess  TokenType = "access"
	TokenRefresh TokenType = "refresh"
)

type Claims struct {
	UserID uuid.UUID      `json:"uid"`
	Email  string         `json:"email"`
	Role   model.UserRole `json:"role"`
	Type   TokenType      `json:"typ"`
	jwt.RegisteredClaims
}

// Issuer firma e verifica token JWT HS256.
type Issuer struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewIssuer(secret string, accessTTL, refreshTTL time.Duration) *Issuer {
	return &Issuer{secret: []byte(secret), accessTTL: accessTTL, refreshTTL: refreshTTL}
}

func (i *Issuer) sign(c Claims) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return t.SignedString(i.secret)
}

func (i *Issuer) IssueAccess(u *model.User) (string, error) {
	now := time.Now()
	return i.sign(Claims{
		UserID: u.ID, Email: u.Email, Role: u.Role, Type: TokenAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   u.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(i.accessTTL)),
		},
	})
}

func (i *Issuer) IssueRefresh(u *model.User) (string, error) {
	now := time.Now()
	return i.sign(Claims{
		UserID: u.ID, Email: u.Email, Role: u.Role, Type: TokenRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   u.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(i.refreshTTL)),
		},
	})
}

func (i *Issuer) Parse(tokenStr string) (*Claims, error) {
	var c Claims
	tok, err := jwt.ParseWithClaims(tokenStr, &c, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return i.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !tok.Valid {
		return nil, errors.New("invalid token")
	}
	return &c, nil
}
