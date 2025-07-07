package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const minSecretKeySize = 32

type JWTMaker struct {
	secretKey string
}

func (maker *JWTMaker) CreateToken(username string, duration time.Duration) (string, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload.GetJWTClaims())

	return token.SignedString([]byte(maker.secretKey))
}

func (maker *JWTMaker) VerifyToken(tokenString string) (*Payload, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, ErrInvalidToken
		}

		return []byte(maker.secretKey), nil
	}
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, keyFunc)
	if err != nil {

		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}

		if errors.Is(err, jwt.ErrTokenUnverifiable) {
			return nil, ErrInvalidToken
		}

		return nil, err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims.GetPayload(), nil
}

func NewJWTMaker(secretKey string) (Maker, error) {
	if len(secretKey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid key size: must be at least %d characters", minSecretKeySize)
	}

	return &JWTMaker{secretKey}, nil
}
