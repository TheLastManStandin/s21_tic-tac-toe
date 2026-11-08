package services

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenUse uint32

const (
	access TokenUse = iota + 1
	refresh
)

type claims struct {
	jwt.RegisteredClaims
	TokenUse TokenUse `json:"token_use"`
}

type JwtProvider struct {
	Key string
}

func NewJwtProvider() (*JwtProvider, error) {
	key := os.Getenv("JWT_KEY")
	if key == "" {
		return nil, fmt.Errorf("No jwt key provided is .env")
	}
	return &JwtProvider{Key: key}, nil
}

func (jp *JwtProvider) GenerateAccessToken(userUUID string) (string, error) {
	userClaims := claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userUUID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		TokenUse: access,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, userClaims)
	return token.SignedString([]byte(jp.Key))
}

func (jp *JwtProvider) GenerateRefreshToken(userUUID string) (string, error) {
	userClaims := claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userUUID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		TokenUse: refresh,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, userClaims)
	return token.SignedString([]byte(jp.Key))
}

func (jp *JwtProvider) validateToken(
	tokenString string,
	expectedUse TokenUse,
) (*claims, error) {
	if jp.Key == "" {
		return nil, fmt.Errorf("JWT key is empty")
	}

	userClaims := &claims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		userClaims,
		func(_ *jwt.Token) (any, error) {
			return []byte(jp.Key), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
	)
	if err != nil {
		return nil, fmt.Errorf("validate JWT: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid JWT")
	}

	if userClaims.TokenUse != expectedUse {
		return nil, fmt.Errorf("unexpected token use")
	}

	if err := uuid.Validate(userClaims.Subject); err != nil {
		return nil, fmt.Errorf("invalid user UUID: %w", err)
	}

	return userClaims, nil
}

func (jp *JwtProvider) ValidateAccessToken(token string) bool {
	_, err := jp.validateToken(token, access)
	return err == nil
}

func (jp *JwtProvider) ValidateRefreshToken(token string) bool {
	_, err := jp.validateToken(token, refresh)
	return err == nil
}
