package domain

import "errors"

var ErrUserAlreadyExists = errors.New("user already exists")
var ErrCannotConnectToGame = errors.New("cannot connect to game")
var ErrGameAccessDenied = errors.New("access denied to game")
var ErrNotYourTurn = errors.New("it is not your turn")
