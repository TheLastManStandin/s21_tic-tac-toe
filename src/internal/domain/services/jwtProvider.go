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

type jwtProvider struct {
}

func (jp *jwtProvider) GenerateAccessToken(user domain.UserInstance) (string, error) {
	userClaims := claims{
		UserUUID: user.UUID,
		Expires:  time.Now().Add(15 * time.Minute).Unix(),
		TokenUse: access,
	}
}
