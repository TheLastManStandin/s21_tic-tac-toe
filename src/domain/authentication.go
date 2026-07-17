package domain

import (
	"datasource"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func VerifyUserPass(username, password string, db datasource.Database) bool {
	user, err := db.GetUserByName(strings.Trim(username, " "))
	if err != nil {
		return false
	}
	if cmperr := bcrypt.CompareHashAndPassword(user.HashPassword, []byte(password)); cmperr != nil {
		return false
	}
	return true
}
