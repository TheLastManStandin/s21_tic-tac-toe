package domain

import (
	"datasource"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func VerifyUserPass(username, password string, db datasource.Database) bool {

}
