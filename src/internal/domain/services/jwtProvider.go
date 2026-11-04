package services

import (
	"tickTackToe/internal/domain"
	"time"
)

type TokenUse uint32

const (
	access  TokenUse = 0
	refresh TokenUse = 1
)

type claims struct {
	UserUUID string   `json:"sub"`
	Expires  int64    `json:"exp"`
	TokenUse TokenUse `json:"token_use"`
}

type JwtProvider struct {
	Key string
}

func NewJwtProvider() *JwtProvider {
	return &JwtProvider{
		Key: ,
	}
}

func (jp *JwtProvider) 

func (jp *JwtProvider) GenerateAccessToken(user domain.UserInstance) (string, error) {
	userClaims := claims{
		UserUUID: user.UUID,
		Expires:  time.Now().Add(15 * time.Minute).Unix(),
		TokenUse: access,
	}
}
