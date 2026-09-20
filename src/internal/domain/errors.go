package domain

import "errors"

var ErrUserAlreadyExists = errors.New("user already exists")
var ErrCannotConnectToGame = errors.New("cannot connect to game")
