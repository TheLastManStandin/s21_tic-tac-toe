package domain

import (
	"datasource"
	"errors"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	database datasource.Database
}

func NewUserService(database datasource.Database) *UserService {
	return &UserService{
		database: database,
	}
}

func (us UserService) Authentication(req http.Request) error {
	username, password, ok := req.BasicAuth()

	user, err := us.database.GetUserByName(strings.Trim(username, " "))

	if err != nil {
		return errors.New("invalid username or password")
	}
	if cmperr := bcrypt.CompareHashAndPassword(user.HashPassword, []byte(password)); cmperr != nil {
		return errors.New("invalid username or password")
	}
	if ok && VerifyUserPass(username, password, us.database) {
		return nil
	}
	return errors.New("invalid username or password")
}
