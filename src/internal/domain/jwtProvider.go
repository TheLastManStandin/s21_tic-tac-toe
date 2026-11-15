package domain

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenUse uint32

const (
	defaultAccessTokenTTL  = 15 * time.Minute
	defaultRefreshTokenTTL = 15 * 24 * time.Hour
)

const (
	access TokenUse = iota + 1
	refresh
)

type claims struct {
	jwt.RegisteredClaims
	TokenUse TokenUse `json:"token_use"`
}

type JwtProvider struct {
	Key        string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewJwtProvider() (*JwtProvider, error) {
	key := os.Getenv("JWT_KEY")
	if key == "" {
		return nil, fmt.Errorf("No jwt key provided is .env")
	}
	accessTTL, err := tokenTTLFromEnv("ACCESS_TOKEN_TTL", defaultAccessTokenTTL)
	if err != nil {
		return nil, err
	}
	refreshTTL, err := tokenTTLFromEnv("REFRESH_TOKEN_TTL", defaultRefreshTokenTTL)
	if err != nil {
		return nil, err
	}
	return &JwtProvider{
		Key:        key,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}, nil
}

func tokenTTLFromEnv(name string, defaultTTL time.Duration) (time.Duration, error) {
	value := os.Getenv(name)
	if value == "" {
		return defaultTTL, nil
	}
	ttl, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", name, err)
	}
	if ttl <= 0 {
		return 0, fmt.Errorf("%s must be positive", name)
	}
	return ttl, nil
}

func (jp *JwtProvider) GenerateAccessToken(userUUID string) (string, error) {
	now := time.Now()
	userClaims := claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userUUID,
			ExpiresAt: jwt.NewNumericDate(now.Add(jp.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		TokenUse: access,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, userClaims)
	return token.SignedString([]byte(jp.Key))
}

func (jp *JwtProvider) GenerateRefreshToken(userUUID string) (string, error) {
	now := time.Now()
	userClaims := claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userUUID,
			ExpiresAt: jwt.NewNumericDate(now.Add(jp.refreshTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
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

func (jp *JwtProvider) ValidateAccessToken(token string) (string, error) {
	claims, err := jp.validateToken(token, access)
	if err != nil {
		return "", err
	}
	return claims.Subject, nil
}

func (jp *JwtProvider) ValidateRefreshToken(token string) (string, error) {
	claims, err := jp.validateToken(token, refresh)
	if err != nil {
		return "", err
	}
	return claims.Subject, nil
}
