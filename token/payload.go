package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token is expired")
)

type Payload struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

func (p *Payload) Valid() error {
	if time.Now().After(p.ExpiredAt) {
		return ErrTokenExpired
	}

	return nil
}

func NewPayload(username string, role string, duration time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	issuedAt := time.Now()
	expiredAt := issuedAt.Add(duration)

	payload := &Payload{
		ID:        tokenID,
		Username:  username,
		Role:      role,
		IssuedAt:  issuedAt,
		ExpiredAt: expiredAt,
	}

	return payload, nil
}

type CustomClaims struct {
	Username string    `json:"username"`
	ID       uuid.UUID `json:"id"`
	Role     string    `json:"role"`
	jwt.RegisteredClaims
}

func (p *Payload) GetJWTClaims() *CustomClaims {
	return &CustomClaims{
		p.Username,
		p.ID,
		p.Role,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(p.ExpiredAt),
			IssuedAt:  jwt.NewNumericDate(p.IssuedAt)},
	}
}

func (c *CustomClaims) GetPayload() *Payload {
	return &Payload{
		ID:        c.ID,
		Username:  c.Username,
		Role:      c.Role,
		IssuedAt:  c.IssuedAt.Time,
		ExpiredAt: c.ExpiresAt.Time,
	}
}
